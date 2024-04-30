package cmd

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strings"

	"log/slog"

	"github.com/spf13/cobra"

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
	rootCmd.AddCommand(hotfix04Cmd)
}

var hotfix04Cmd = &cobra.Command{
	Use:   "hotfix04",
	Short: "Download private files from the old workery to a local temporary directory",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		defaultLogger := slog.Default()
		ctx := context.Background()
		cfg := config.New()
		mc := mongodb.NewStorage(cfg)
		s3 := s3storage.NewStorage(cfg, defaultLogger)
		oldS3 := s3storage.NewStorageWithCustom(defaultLogger, cfg.OldAWS.Endpoint, cfg.OldAWS.Region, cfg.OldAWS.AccessKey, cfg.OldAWS.SecretKey, cfg.OldAWS.BucketName, false)
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

		RunHotfix04(cfg, defaultLogger, ppc, lpc, aStorer, uStorer, cStorer, asStorer, oStorer, sStorer, tenant, s3, oldS3)
	},
}

func RunHotfix04(
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

	// STEP 2: Fetch all the s3objects.
	allOldS3Objects, err := oldS3.ListAllObjects(context.Background())
	if err != nil {
		logger.Error("list all objects", slog.Any("err", err))
		panic("list all objects")
	}

	// STEP 3: Iterate through all the s3objects.
	for _, obj := range allOldS3Objects.Contents {
		// Get the key.
		objectKey := *obj.Key

		//
		// DEVELOPERS NOTE:
		// If this code block runs then the private file gets imported.
		// The following code will save to local directory.
		//

		// Get the filename.
		segements := strings.Split(objectKey, "/")
		fileName := segements[len(segements)-1]

		// Get the directory to save.
		directory := "./static/" + fileName

		// Save and get the filepath.
		localFilePath, err := oldS3.DownloadToLocalfile(context.Background(), objectKey, directory)
		if err != nil {
			logger.Error("download to local file error", slog.Any("err", err))
			logger.Warn("skipping file to download...")
			continue // Skip this current loop iteration to another loop.
		}

		// For debugging purposes only.
		log.Println("---->", localFilePath, "<----")
	}

	fmt.Println("Finished importing private images")
}
