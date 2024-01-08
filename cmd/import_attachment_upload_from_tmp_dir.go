package cmd

import (
	"context"
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"log/slog"
	"strings"

	"github.com/spf13/cobra"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/over55/workery-cli/adapter/storage/mongodb"
	"github.com/over55/workery-cli/adapter/storage/postgres"
	s3storage "github.com/over55/workery-cli/adapter/storage/s3"
	a_ds "github.com/over55/workery-cli/app/associate/datastore"
	pi_ds "github.com/over55/workery-cli/app/attachment/datastore"
	c_ds "github.com/over55/workery-cli/app/customer/datastore"
	o_ds "github.com/over55/workery-cli/app/order/datastore"
	s_ds "github.com/over55/workery-cli/app/staff/datastore"
	tenant_ds "github.com/over55/workery-cli/app/tenant/datastore"
	user_ds "github.com/over55/workery-cli/app/user/datastore"
	"github.com/over55/workery-cli/config"
)

func init() {
	rootCmd.AddCommand(importAttachmentUploadCmd)
}

var importAttachmentUploadCmd = &cobra.Command{
	Use:   "import_attachment_upload_from_tmp_dir",
	Short: "Download private files from the old workery to a local temporary directory",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		defaultLogger := slog.Default()
		ctx := context.Background()
		cfg := config.New()
		mc := mongodb.NewStorage(cfg)
		s3 := s3storage.NewStorage(cfg, defaultLogger)
		oldS3 := s3storage.NewStorageWithCustom(
			defaultLogger,
			cfg.OldAWS.Endpoint,
			cfg.OldAWS.Region,
			cfg.OldAWS.AccessKey,
			cfg.OldAWS.SecretKey,
			cfg.OldAWS.BucketName,
		)
		ppc := postgres.NewStorage(cfg, cfg.PostgresDB.DatabasePublicSchemaName)
		lpc := postgres.NewStorage(cfg, cfg.PostgresDB.DatabaseLondonSchemaName)

		tStorer := tenant_ds.NewDatastore(cfg, defaultLogger, mc)
		aStorer := pi_ds.NewDatastore(cfg, defaultLogger, mc)
		uStorer := user_ds.NewDatastore(cfg, defaultLogger, mc)
		cStorer := c_ds.NewDatastore(cfg, defaultLogger, mc)
		asStorer := a_ds.NewDatastore(cfg, defaultLogger, mc)
		oStorer := o_ds.NewDatastore(cfg, defaultLogger, mc)
		sStorer := s_ds.NewDatastore(cfg, defaultLogger, mc)

		tenant, err := tStorer.GetBySchemaName(ctx, cfg.PostgresDB.DatabaseLondonSchemaName)
		if err != nil {
			defaultLogger.Error("get schema name", slog.Any("err", err))
			panic("get schema name")
		}

		RunImportAttachmentUpload(cfg, defaultLogger, ppc, lpc, aStorer, uStorer, cStorer, asStorer, oStorer, sStorer, tenant, s3, oldS3)
	},
}

func RunImportAttachmentUpload(
	cfg *config.Conf,
	logger *slog.Logger,
	private *sql.DB,
	london *sql.DB,
	aStorer pi_ds.AttachmentStorer,
	uStorer user_ds.UserStorer,
	cStorer c_ds.CustomerStorer,
	asStorer a_ds.AssociateStorer,
	oStorer o_ds.OrderStorer,
	sStorer s_ds.StaffStorer,
	tenant *tenant_ds.Tenant,
	s3 s3storage.S3Storager,
	oldS3 s3storage.S3Storager,
) {
	fmt.Println("Beginning uploading attachments")

	f := &pi_ds.AttachmentListFilter{
		Cursor:    primitive.NilObjectID,
		PageSize:  1_000_000,
		SortField: "_id",
		SortOrder: 1,
	}

	attachments, err := aStorer.ListByFilter(context.Background(), f)
	if err != nil {
		logger.Error("get by old id", slog.Any("err", err))
		panic("get by old id")
	}

	logger.Debug("fetched", slog.Any("results", attachments.Results))

	for _, attachment := range attachments.Results {
		importAttachmentUpload(context.Background(), logger, s3, aStorer, attachment)
	}

	fmt.Println("Finished uploading attachments")
}

func importAttachmentUpload(
	ctx context.Context,
	logger *slog.Logger,
	s3 s3storage.S3Storager,
	aStorer pi_ds.AttachmentStorer,
	attachment *pi_ds.Attachment,
) {
	newObjectKey := "tenant/" + attachment.TenantID.Hex() + "/private/uploads" + strings.Replace(attachment.ObjectKey, "./static", "", 1)

	logger.Debug("proceeding to upload", slog.String("object_key", newObjectKey))

	//
	// Open and read file.
	//

	// Read the file contents into a []byte slice
	fileContent, err := ioutil.ReadFile(attachment.ObjectKey)
	if err != nil {
		log.Fatal(err)
	}

	//
	// Upload the content to S3.
	//

	if err := s3.UploadContent(ctx, newObjectKey, fileContent); err != nil {
		log.Fatal("UploadBinToS3", err)
	}
	// Update the private file in the database.
	attachment.ObjectKey = newObjectKey
	if err := aStorer.UpdateByID(context.Background(), attachment); err != nil {
		log.Fatal("pfr.UpdateById:", err)
	}

	fmt.Println("Uploaded Attachment ID#", attachment.ID)
}
