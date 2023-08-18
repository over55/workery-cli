package cmd

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/spf13/cobra"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/exp/slog"
	"gopkg.in/guregu/null.v4"

	"github.com/over55/workery-cli/adapter/storage/mongodb"
	"github.com/over55/workery-cli/adapter/storage/postgres"
	as_ds "github.com/over55/workery-cli/app/activitysheet/datastore"
	a_ds "github.com/over55/workery-cli/app/associate/datastore"
	o_ds "github.com/over55/workery-cli/app/order/datastore"
	tenant_ds "github.com/over55/workery-cli/app/tenant/datastore"
	user_ds "github.com/over55/workery-cli/app/user/datastore"
	"github.com/over55/workery-cli/config"
)

func init() {
	rootCmd.AddCommand(importActivitySheetCmd)
}

var importActivitySheetCmd = &cobra.Command{
	Use:   "import_activity_sheet",
	Short: "Import the activity sheets from old database",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		cfg := config.New()
		mc := mongodb.NewStorage(cfg)
		ppc := postgres.NewStorage(cfg, cfg.PostgresDB.DatabasePublicSchemaName)
		lpc := postgres.NewStorage(cfg, cfg.PostgresDB.DatabaseLondonSchemaName)

		defaultLogger := slog.Default()

		tenantStorer := tenant_ds.NewDatastore(cfg, defaultLogger, mc)
		asStorer := as_ds.NewDatastore(cfg, defaultLogger, mc)
		uStorer := user_ds.NewDatastore(cfg, defaultLogger, mc)
		aStorer := a_ds.NewDatastore(cfg, defaultLogger, mc)
		oStorer := o_ds.NewDatastore(cfg, defaultLogger, mc)

		tenant, err := tenantStorer.GetBySchemaName(ctx, cfg.PostgresDB.DatabaseLondonSchemaName)
		if err != nil {
			log.Fatal(err)
		}

		RunImportActivitySheet(cfg, ppc, lpc, aStorer, asStorer, uStorer, oStorer, tenant)
	},
}

func RunImportActivitySheet(cfg *config.Conf, public *sql.DB, london *sql.DB, aStorer a_ds.AssociateStorer, asStorer as_ds.ActivitySheetStorer, uStorer user_ds.UserStorer, oStorer o_ds.OrderStorer, tenant *tenant_ds.Tenant) {
	fmt.Println("Beginning importing activity sheets")
	data, err := ListAllActivitySheetItems(london)
	if err != nil {
		log.Fatal(err)
	}

	for _, datum := range data {
		importActivitySheet(context.Background(), aStorer, asStorer, uStorer, oStorer, tenant, datum)
	}
	fmt.Println("Finished importing activity sheets")
}

type OldUActivitySheetItem struct {
	ID           uint64      `json:"id"`
	Comment      string      `json:"comment"`
	CreatedAt    time.Time   `json:"created_at"`
	CreatedFrom  null.String `json:"created_from"`
	CreatedByID  null.Int    `json:"created_by_id"`
	AssociateID  uint64      `json:"associate_id"`
	JobID        null.Int    `json:"job_id"`
	State        string      `json:"state"`
	OngoingJobID null.Int    `json:"ongoing_job_id"`
}

func ListAllActivitySheetItems(db *sql.DB) ([]*OldUActivitySheetItem, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `
	SELECT
	    id, comment, created_at, created_from, created_by_id, associate_id, job_id, state, ongoing_job_id
	FROM
	    london.workery_activity_sheet_items
	ORDER BY
	    id
	ASC
	`
	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}

	var arr []*OldUActivitySheetItem
	defer rows.Close()
	for rows.Next() {
		m := new(OldUActivitySheetItem)
		err = rows.Scan(
			&m.ID,
			&m.Comment,
			&m.CreatedAt,
			&m.CreatedFrom,
			&m.CreatedByID,
			&m.AssociateID,
			&m.JobID,
			&m.State,
			&m.OngoingJobID,
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

func importActivitySheet(ctx context.Context, aStorer a_ds.AssociateStorer, asStorer as_ds.ActivitySheetStorer, uStorer user_ds.UserStorer, oStorer o_ds.OrderStorer, tenant *tenant_ds.Tenant, asi *OldUActivitySheetItem) {
	//
	// Compile our `state`.
	//

	var state int8 = 1
	if asi.State == "pending" {
		state = as_ds.ActivitySheetStatusPending
	} else if asi.State == "accepted" {
		state = as_ds.ActivitySheetStatusAccepted
	} else if asi.State == "declined" {
		state = as_ds.ActivitySheetStatusDeclined
	}

	//
	// Get `createdByID` and `createdByName` values.
	//

	var createdByID primitive.ObjectID = primitive.NilObjectID
	var createdByName string
	if asi.CreatedByID.ValueOrZero() > 0 {
		user, err := uStorer.GetByOldID(ctx, uint64(asi.CreatedByID.ValueOrZero()))
		if err != nil {
			log.Fatal("ur.GetByOldID", err)
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
	if asi.CreatedByID.ValueOrZero() > 0 {
		user, err := uStorer.GetByOldID(ctx, uint64(asi.CreatedByID.ValueOrZero()))
		if err != nil {
			log.Fatal("ur.GetByOldID", err)
		}
		if user != nil {
			modifiedByID = user.ID
			modifiedByName = user.Name
		}
	}

	//
	// Lookup related.
	//

	associate, err := aStorer.GetByOldID(ctx, asi.AssociateID)
	if err != nil {
		log.Fatal(err)
	}
	if associate == nil {
		log.Fatal("associate does not exist")
	}

	var orderID primitive.ObjectID = primitive.NilObjectID
	order, err := oStorer.GetByOldID(ctx, uint64(asi.JobID.ValueOrZero()))
	if err != nil {
		log.Fatal(err)
	}
	if order != nil {
		orderID = order.ID
	}

	m := &as_ds.ActivitySheet{
		ID:                    primitive.NewObjectID(),
		TenantID:              tenant.ID,
		Comment:               asi.Comment,
		CreatedAt:             asi.CreatedAt,
		CreatedByUserID:       createdByID,
		CreatedByUserName:     createdByName,
		CreatedFromIPAddress:  "",
		ModifiedAt:            time.Now(),
		ModifiedByUserID:      modifiedByID,
		ModifiedByUserName:    modifiedByName,
		ModifiedFromIPAddress: "",
		AssociateID:           associate.ID,
		AssociateName:         associate.Name,
		AssociateLexicalName:  associate.LexicalName,
		OrderID:               orderID,
		Status:                state,
		TypeOf:                associate.TypeOf,
		OldID:                 asi.ID,
	}

	if err := asStorer.Create(ctx, m); err != nil {
		log.Panic(err)
	}
	fmt.Println("Imported ActivitySheet ID#", m.ID)
}
