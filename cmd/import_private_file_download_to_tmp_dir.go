package cmd

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/spf13/cobra"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/exp/slog"
	"gopkg.in/guregu/null.v4"

	"github.com/over55/workery-cli/adapter/storage/mongodb"
	"github.com/over55/workery-cli/adapter/storage/postgres"
	s3storage "github.com/over55/workery-cli/adapter/storage/s3"
	pi_ds "github.com/over55/workery-cli/app/privateimage/datastore"
	tenant_ds "github.com/over55/workery-cli/app/tenant/datastore"
	"github.com/over55/workery-cli/config"
)

func init() {
	rootCmd.AddCommand(importPrivateImageCmd)
}

var importPrivateImageCmd = &cobra.Command{
	Use:   "import_private_file_download_to_tmp_dir",
	Short: "Download private files from the old workery to a local temporary directory",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		defaultLogger := slog.Default()
		ctx := context.Background()
		cfg := config.New()
		mc := mongodb.NewStorage(cfg)
		s3 := s3storage.NewStorage(cfg, defaultLogger)
		oldS3 := s3storage.NewStorageWithCustom(defaultLogger, cfg.OldAWS.Endpoint, cfg.OldAWS.Region, cfg.OldAWS.AccessKey, cfg.OldAWS.SecretKey, cfg.OldAWS.BucketName)
		ppc := postgres.NewStorage(cfg, cfg.PostgresDB.DatabasePublicSchemaName)
		lpc := postgres.NewStorage(cfg, cfg.PostgresDB.DatabaseLondonSchemaName)

		tenantStorer := tenant_ds.NewDatastore(cfg, defaultLogger, mc)
		irStorer := pi_ds.NewDatastore(cfg, defaultLogger, mc)

		tenant, err := tenantStorer.GetBySchemaName(ctx, cfg.PostgresDB.DatabaseLondonSchemaName)
		if err != nil {
			log.Fatal(err)
		}

		RunImportPrivateImage(cfg, ppc, lpc, irStorer, tenant, s3, oldS3)
	},
}

func RunImportPrivateImage(
	cfg *config.Conf,
	private *sql.DB,
	london *sql.DB,
	irStorer pi_ds.PrivateImageStorer,
	tenant *tenant_ds.Tenant,
	s3 s3storage.S3Storager,
	oldS3 s3storage.S3Storager,
) {
	fmt.Println("Beginning importing private images")

	// STEP 1: Fetch old database files.
	oldData, err := ListAllOldPrivateFiles(london)
	if err != nil {
		log.Fatal(err)
	}

	// STEP 2: Fetch all the s3objects.
	allOldS3Objects, err := oldS3.ListAllObjects(context.Background())
	if err != nil {
		log.Fatal(err)
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
					log.Fatal(err)
				}

				// For debugging purposes only.
				log.Println("---->", localFilePath, "<----")

				//
				// Lookup related files.
				//

				m := &pi_ds.PrivateImage{

					ID:                    primitive.NewObjectID(),
					TenantID:              tenant.ID,
					ObjectKey:             localFilePath,
					Title:                 oldDatum.Title,
					Description:           oldDatum.Description,
					CreatedAt:             oldDatum.CreatedAt,
					CreatedByUserID:       primitive.NilObjectID,
					CreatedByUserName:     "",
					CreatedFromIPAddress:  oldDatum.CreatedFrom.ValueOrZero(),
					ModifiedAt:            oldDatum.LastModifiedAt,
					ModifiedByUserID:      primitive.NilObjectID,
					ModifiedByUserName:    "",
					ModifiedFromIPAddress: oldDatum.LastModifiedFrom.ValueOrZero(),
					AssociateID:           primitive.NilObjectID,
					CustomerID:            primitive.NilObjectID,
					StaffID:               primitive.NilObjectID,
					OrderID:               primitive.NilObjectID,
					Status:                1,
					OldID:                 oldDatum.ID,
				}

				if err := irStorer.Create(context.Background(), m); err != nil {
					log.Panic(err)
				}
				fmt.Println("Imported PrivateImage ID#", m.ID)

				return
			}
		}
	}

	// for _, datum := range data {
	// 	importPrivateImage(context.Background(), irStorer, tenant, datum, s3, oldS3, allOldS3Objects)
	// }
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

func importPrivateImage(
	ctx context.Context,
	irStorer pi_ds.PrivateImageStorer,
	tenant *tenant_ds.Tenant,
	oldFile *OldPrivateFile,
	s3 s3storage.S3Storager,
	oldS3 s3storage.S3Storager,
	allS3Objects *s3.ListObjectsOutput,
) {
	// var state int8 = 1
	// if t.IsArchived == true {
	// 	state = 2
	// }
	//
	// m := &pi_ds.PrivateImage{
	// 	OldID:       t.ID,
	// 	ID:          primitive.NewObjectID(),
	// 	Text:        t.Text,
	// 	Description: t.Description,
	// 	Status:      state,
	// 	TenantID:    tenant.ID,
	// }
	// err := irStorer.Create(ctx, m)
	// if err != nil {
	// 	log.Panic(err)
	// }
	// fmt.Println("Imported PrivateImage ID#", m.ID)
}
