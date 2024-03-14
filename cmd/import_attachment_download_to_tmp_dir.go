package cmd

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	"log/slog"

	"github.com/spf13/cobra"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"gopkg.in/guregu/null.v4"

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
	rootCmd.AddCommand(importAttachmentDownloadCmd)
}

var importAttachmentDownloadCmd = &cobra.Command{
	Use:   "import_attachment_download_to_tmp_dir",
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

		RunImportAttachmentDownload(cfg, defaultLogger, ppc, lpc, aStorer, uStorer, cStorer, asStorer, oStorer, sStorer, tenant, s3, oldS3)
	},
}

func RunImportAttachmentDownload(
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
	fmt.Println("Beginning importing private images")

	// STEP 1: Fetch old database files.
	oldData, err := ListAllOldPrivateFiles(london)
	if err != nil {
		logger.Error("list all old private files", slog.Any("err", err))
		panic("list all old private files")
	}

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

		// STEP 4: Iterate through all the old database files.
		for _, oldDatum := range oldData {

			// Get the filename.
			segements := strings.Split(oldDatum.DataFile, "/")
			oldFileName := segements[len(segements)-1]

			// Check to see if the filenames match.
			match := strings.Contains(objectKey, oldFileName)

			// STEP 5:
			// If a match happens then it means we have found the ACTUAL KEY in the
			// s3 objects inside the bucket.
			if match == true {
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

				// // For debugging purposes only.
				// log.Println("---->", localFilePath, "<----")

				//
				// Lookup related files and import into database.
				//

				importAttachment(context.Background(), logger, tenant, localFilePath, oldDatum, aStorer, uStorer, asStorer, cStorer, oStorer, sStorer)
			}
		}
	}

	fmt.Println("Finished importing private images")
}

type OldPrivateFile struct {
	ID                       uint64      `json:"id"`
	DataFile                 string      `json:"data_file"`
	Title                    string      `json:"title"`
	Description              string      `json:"description"`
	IsArchived               bool        `json:"is_archived"`
	IndexedText              null.String `json:"indexed_text"`
	CreatedAt                time.Time   `json:"created_at"`
	CreatedFrom              null.String `json:"created_from"`
	CreatedFromIsPublic      bool        `json:"created_from_is_public"`
	CreatedByID              null.Int    `json:"created_by_id"`
	LastModifiedAt           time.Time   `json:"last_modified_at"`
	LastModifiedFrom         null.String `json:"last_modified_from"`
	LastModifiedFromIsPublic bool        `json:"last_modified_from_is_public"`
	LastModifiedByID         null.Int    `json:"last_modified_by_id"`
	AssociateID              null.Int    `json:"associate_id"`
	CustomerID               null.Int    `json:"customer_id"`
	PartnerID                null.Int    `json:"partner_id"`
	StaffID                  null.Int    `json:"staff_id"`
	WorkOrderID              null.Int    `json:"work_order_id"`
}

func ListAllOldPrivateFiles(db *sql.DB) ([]*OldPrivateFile, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `
	SELECT
	    id, data_file, title, description, is_archived, indexed_text, created_at,
		created_from, created_from_is_public, created_by_id, last_modified_at,
		last_modified_from, last_modified_from_is_public, last_modified_by_id,
		associate_id, customer_id, partner_id, staff_id, work_order_id
	FROM
	    london.workery_private_file_uploads
	ORDER BY
	    id
	ASC
	`
	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}

	var arr []*OldPrivateFile
	defer rows.Close()
	for rows.Next() {
		m := new(OldPrivateFile)
		err = rows.Scan(
			&m.ID,
			&m.DataFile,
			&m.Title,
			&m.Description,
			&m.IsArchived,
			&m.IndexedText,
			&m.CreatedAt,
			&m.CreatedFrom,
			&m.CreatedFromIsPublic,
			&m.CreatedByID,
			&m.LastModifiedAt,
			&m.LastModifiedFrom,
			&m.LastModifiedFromIsPublic,
			&m.LastModifiedByID,
			&m.AssociateID,
			&m.CustomerID,
			&m.PartnerID,
			&m.StaffID,
			&m.WorkOrderID,
		)
		if err != nil {
			panic(err)
		}
		arr = append(arr, m)
	}
	err = rows.Err()
	if err != nil {
		panic(err)
	}
	return arr, err
}

func importAttachment(
	ctx context.Context,
	logger *slog.Logger,
	tenant *tenant_ds.Tenant,
	localFilePath string,
	oldDatum *OldPrivateFile,
	aStorer pi_ds.AttachmentStorer,
	uStorer user_ds.UserStorer,
	asStorer a_ds.AssociateStorer,
	cStorer c_ds.CustomerStorer,
	oStorer o_ds.OrderStorer,
	sStorer s_ds.StaffStorer,
) {
	//
	// Initial variables.
	//

	var typeOf int8 = 0

	//
	// Get `createdByID` and `createdByName` values.
	//

	var createdByID primitive.ObjectID = primitive.NilObjectID
	var createdByName string
	if oldDatum.CreatedByID.ValueOrZero() > 0 {
		user, err := uStorer.GetByPublicID(ctx, uint64(oldDatum.CreatedByID.ValueOrZero()))
		if err != nil {
			logger.Error("get by old id", slog.Any("err", err))
			panic("get by old id")
		}
		if user != nil {
			createdByID = user.ID
			createdByName = user.Name
		}
	}

	//
	// Get `modifiedByID` and `modifiedByName` values.
	//

	var modifiedByID primitive.ObjectID = primitive.NilObjectID
	var modifiedByName string
	if oldDatum.CreatedByID.ValueOrZero() > 0 {
		user, err := uStorer.GetByPublicID(ctx, uint64(oldDatum.CreatedByID.ValueOrZero()))
		if err != nil {
			logger.Error("get by old id", slog.Any("err", err))
			panic("get by old id")
		}
		if user != nil {
			modifiedByID = user.ID
			modifiedByName = user.Name
		}
	}

	//
	// Customer
	//

	var customerID primitive.ObjectID = primitive.NilObjectID
	var customerName = ""
	if !oldDatum.CustomerID.IsZero() {
		customer, err := cStorer.GetByPublicID(ctx, uint64(oldDatum.CustomerID.ValueOrZero()))
		if err != nil {
			logger.Error("get by old id", slog.Any("err", err))
			panic("get by old id")
		}
		if customer != nil {
			customerID = customer.ID
			customerName = customer.Name
			typeOf = pi_ds.AttachmentTypeCustomer
		}
	}

	//
	// Associate
	//

	var associateID primitive.ObjectID = primitive.NilObjectID
	var associateName string = ""
	if !oldDatum.AssociateID.IsZero() {
		associate, err := asStorer.GetByPublicID(ctx, uint64(oldDatum.AssociateID.ValueOrZero()))
		if err != nil {
			logger.Error("get by old id", slog.Any("err", err))
			panic("get by old id")
		}
		if associate != nil {
			associateID = associate.ID
			associateName = associate.Name
			typeOf = pi_ds.AttachmentTypeAssociate
		}
	}

	//
	// Order
	//

	var orderID primitive.ObjectID = primitive.NilObjectID
	if !oldDatum.WorkOrderID.IsZero() {
		order, err := oStorer.GetByWJID(ctx, uint64(oldDatum.WorkOrderID.ValueOrZero()))
		if err != nil {
			log.Fatal(err)
		}
		if order != nil {
			orderID = order.ID
			typeOf = pi_ds.AttachmentTypeOrder
		}
	}

	//
	// Staff
	//

	var staffID primitive.ObjectID = primitive.NilObjectID
	var staffName string = ""
	if !oldDatum.StaffID.IsZero() {
		staff, err := sStorer.GetByPublicID(ctx, uint64(oldDatum.StaffID.ValueOrZero()))
		if err != nil {
			logger.Error("get by old id", slog.Any("err", err))
			panic("get by old id")
		}
		if staff != nil {
			staffID = staff.ID
			staffName = staff.Name
			typeOf = pi_ds.AttachmentTypeStaff
		}
	}

	//
	// Save the database record.
	//

	m := &pi_ds.Attachment{
		ID:                    primitive.NewObjectID(),
		TenantID:              tenant.ID,
		ObjectKey:             localFilePath,
		Title:                 oldDatum.Title,
		Description:           oldDatum.Description,
		CreatedAt:             oldDatum.CreatedAt,
		CreatedByUserID:       createdByID,
		CreatedByUserName:     createdByName,
		CreatedFromIPAddress:  oldDatum.CreatedFrom.ValueOrZero(),
		ModifiedAt:            oldDatum.LastModifiedAt,
		ModifiedByUserID:      modifiedByID,
		ModifiedByUserName:    modifiedByName,
		ModifiedFromIPAddress: oldDatum.LastModifiedFrom.ValueOrZero(),
		CustomerID:            customerID,
		CustomerName:          customerName,
		AssociateID:           associateID,
		AssociateName:         associateName,
		StaffID:               staffID,
		StaffName:             staffName,
		OrderID:               orderID,
		Status:                1,
		Type:                  typeOf,
		PublicID:              oldDatum.ID,
	}

	if err := aStorer.Create(context.Background(), m); err != nil {
		logger.Error("create", slog.Any("err", err))
		panic("create")
	}
	fmt.Println("Imported Attachment ID#", m.ID)
}
