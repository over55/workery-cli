package cmd

import (
	"context"
	"database/sql"
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"strings"

	"log/slog"

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
	rootCmd.AddCommand(hotfix03Cmd)
}

var hotfix03Cmd = &cobra.Command{
	Use:   "hotfix03",
	Short: "Execute hotfix03",
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

		RunHotfix03(cfg, defaultLogger, ppc, lpc, aStorer, uStorer, cStorer, asStorer, oStorer, sStorer, tenant, s3, oldS3)
	},
}

func RunHotfix03(
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
	fmt.Println("Beginning hotfix03")

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

	f, e := os.Create("./attachments.csv")
	if e != nil {
		fmt.Println(e)
	}

	writer := csv.NewWriter(f)
	var data = [][]string{
		{
			"ID",
			"Customer ID",
			"Customer Name",
			"Associate ID",
			"Associate Name",
			"Order ID",
			"Staff ID",
			"Staff Name",
			"Title",
			"Description",
			"Creation",
			"ObjectKey",
		},
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
				// segements := strings.Split(objectKey, "/")
				// fileName := segements[len(segements)-1]

				// // Get the directory to save.
				// directory := "./static/" + fileName

				//
				// Lookup related files and import into database.
				//

				data = executeExportAttachmentsCSV(context.Background(), logger, tenant, objectKey, oldDatum, aStorer, uStorer, asStorer, cStorer, oStorer, sStorer, data)
			}
		}
	}

	e = writer.WriteAll(data)
	if e != nil {
		fmt.Println(e)
	}

	fmt.Println("Finished hotfix03")
}

func executeExportAttachmentsCSV(
	ctx context.Context,
	logger *slog.Logger,
	tenant *tenant_ds.Tenant,
	objectKey string,
	oldDatum *OldPrivateFile,
	aStorer pi_ds.AttachmentStorer,
	uStorer user_ds.UserStorer,
	asStorer a_ds.AssociateStorer,
	cStorer c_ds.CustomerStorer,
	oStorer o_ds.OrderStorer,
	sStorer s_ds.StaffStorer,
	data [][]string,
) [][]string {
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
		ObjectKey:             objectKey,
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

	data = append(data, []string{
		m.ID.Hex(),
		m.CustomerID.Hex(),
		m.CustomerName,
		m.AssociateID.Hex(),
		m.AssociateName,
		m.OrderID.Hex(),
		m.StaffID.Hex(),
		m.StaffName,
		m.Title,
		m.Description,
		m.CreatedAt.Format("2006-01-02 15:04:05"),
		objectKey,
	})

	return data
}
