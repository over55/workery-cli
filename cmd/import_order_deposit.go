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
	a_ds "github.com/over55/workery-cli/app/associate/datastore"
	c_ds "github.com/over55/workery-cli/app/customer/datastore"
	order_ds "github.com/over55/workery-cli/app/order/datastore"
	tenant_ds "github.com/over55/workery-cli/app/tenant/datastore"
	user_ds "github.com/over55/workery-cli/app/user/datastore"
	"github.com/over55/workery-cli/config"
)

func init() {
	rootCmd.AddCommand(importOrderDepositCmd)
}

var importOrderDepositCmd = &cobra.Command{
	Use:   "import_order_deposit",
	Short: "Import the order deposits from old database",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		cfg := config.New()
		mc := mongodb.NewStorage(cfg)
		ppc := postgres.NewStorage(cfg, cfg.PostgresDB.DatabasePublicSchemaName)
		lpc := postgres.NewStorage(cfg, cfg.PostgresDB.DatabaseLondonSchemaName)

		defaultLogger := slog.Default()

		tenantStorer := tenant_ds.NewDatastore(cfg, defaultLogger, mc)
		oStorer := order_ds.NewDatastore(cfg, defaultLogger, mc)
		uStorer := user_ds.NewDatastore(cfg, defaultLogger, mc)
		cStorer := c_ds.NewDatastore(cfg, defaultLogger, mc)
		aStorer := a_ds.NewDatastore(cfg, defaultLogger, mc)

		tenant, err := tenantStorer.GetBySchemaName(ctx, cfg.PostgresDB.DatabaseLondonSchemaName)
		if err != nil {
			log.Fatal(err)
		}

		RunImportOrderDeposit(cfg, ppc, lpc, oStorer, uStorer, aStorer, cStorer, tenant)
	},
}

func RunImportOrderDeposit(
	cfg *config.Conf,
	public *sql.DB,
	london *sql.DB,
	oStorer order_ds.OrderStorer,
	uStorer user_ds.UserStorer,
	aStorer a_ds.AssociateStorer,
	cStorer c_ds.CustomerStorer,
	tenant *tenant_ds.Tenant,
) {
	fmt.Println("Beginning importing order deposits")
	data, err := ListAllWorkOrderDeposits(london)
	if err != nil {
		log.Fatal(err)
	}

	for _, datum := range data {
		importOrderDeposit(context.Background(), oStorer, uStorer, aStorer, cStorer, tenant, datum)
	}
	fmt.Println("Finished importing orders deposits")
}

type OldUWorkOrderDeposit struct {
	ID                       uint64      `json:"id"`
	PaidAt                   null.Time   `json:"paid_at"`
	DepositMethod            int8        `json:"deposit_method"`
	PaidTo                   null.Int    `json:"paid_to"`
	AmountCurrency           string      `json:"amount_currency"`
	Amount                   float64     `json:"amount"`
	PaidFor                  int8        `json:"paid_for"`
	IsArchived               bool        `json:"is_archived"`
	CreatedAt                time.Time   `json:"created_at"`
	LastModifiedAt           time.Time   `json:"last_modified_at"`
	CreatedByID              null.Int    `json:"created_by_id"`
	LastModifiedByID         null.Int    `json:"last_modified_by_id"`
	OrderID                  uint64      `json:"order_id"`
	CreatedFrom              null.String `json:"created_from"`
	CreatedFromIsPublic      bool        `json:"created_from_is_public"`
	LastModifiedFrom         null.String `json:"last_modified_from"`
	LastModifiedFromIsPublic bool        `json:"last_modified_from_is_public"`
}

/*
 inet,
 boolean NOT NULL,
*/

func ListAllWorkOrderDeposits(db *sql.DB) ([]*OldUWorkOrderDeposit, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `
	SELECT
	    id, paid_at, deposit_method, paid_to, amount_currency, amount, paid_for,
		is_archived, created_at, last_modified_at, created_by_id, last_modified_by_id,
		order_id, created_from, created_from_is_public, last_modified_from,
		last_modified_from_is_public
	FROM
	    workery_work_order_deposits
	ORDER BY
	    id
	ASC
	`
	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}

	var arr []*OldUWorkOrderDeposit
	defer rows.Close()
	for rows.Next() {
		m := new(OldUWorkOrderDeposit)
		err = rows.Scan(
			&m.ID,
			&m.PaidAt,
			&m.DepositMethod,
			&m.PaidTo,
			&m.AmountCurrency,
			&m.Amount,
			&m.PaidFor,
			&m.IsArchived,
			&m.CreatedAt,
			&m.LastModifiedAt,
			&m.CreatedByID,
			&m.LastModifiedByID,
			&m.OrderID,
			&m.CreatedFrom,
			&m.CreatedFromIsPublic,
			&m.LastModifiedFrom,
			&m.LastModifiedFromIsPublic,
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

func importOrderDeposit(
	ctx context.Context,
	oStorer order_ds.OrderStorer,
	uStorer user_ds.UserStorer,
	aStorer a_ds.AssociateStorer,
	cStorer c_ds.CustomerStorer,
	tenant *tenant_ds.Tenant,
	od *OldUWorkOrderDeposit) {

	//
	// Status
	//

	var state int8 = 1
	if od.IsArchived == true {
		state = 2
	}

	//
	// Get our `OrderId` value.
	//

	order, err := oStorer.GetByWJID(ctx, od.OrderID)
	if err != nil {
		log.Fatal(err)
	}
	if order == nil {
		log.Fatal("order does not exist")
	}

	//
	// Get created by
	//

	var createdByUserID primitive.ObjectID = primitive.NilObjectID
	var createdByUserName string
	createdByUser, _ := uStorer.GetByOldID(ctx, uint64(od.CreatedByID.ValueOrZero()))
	if createdByUser != nil {
		createdByUserID = createdByUser.ID
		createdByUserName = createdByUser.Name
	}

	//
	// Get modified by
	//

	var modifiedByUserID primitive.ObjectID = primitive.NilObjectID
	var modifiedByUserName string
	modifiedByUser, _ := uStorer.GetByOldID(ctx, uint64(od.CreatedByID.ValueOrZero()))
	if modifiedByUser != nil {
		modifiedByUserID = modifiedByUser.ID
		modifiedByUserName = modifiedByUser.Name
	}

	//
	// Create deposit record.
	//

	deposit := &order_ds.OrderDeposit{
		OldID:                 od.ID,
		OrderWJID:             order.WJID,
		OrderID:               order.ID,
		ID:                    primitive.NewObjectID(),
		TenantID:              tenant.ID,
		PaidAt:                od.PaidAt.ValueOrZero(),
		DepositMethod:         od.DepositMethod,
		PaidTo:                int8(od.PaidTo.ValueOrZero()),
		Currency:              od.AmountCurrency,
		Amount:                od.Amount,
		PaidFor:               od.PaidFor,
		CreatedAt:             od.CreatedAt,
		CreatedByUserID:       createdByUserID,
		CreatedByUserName:     createdByUserName,
		ModifiedByUserID:      modifiedByUserID,
		ModifiedByUserName:    modifiedByUserName,
		ModifiedAt:            od.LastModifiedAt,
		CreatedFromIPAddress:  od.CreatedFrom.ValueOrZero(),
		ModifiedFromIPAddress: od.LastModifiedFrom.ValueOrZero(),
		Status:                state,
	}

	// Append comments to customer details.
	order.Deposits = append(order.Deposits, deposit)

	if err := oStorer.UpdateByID(ctx, order); err != nil {
		log.Fatal(err)
	}

	fmt.Println("Imported Order Deposits ID#", deposit.ID, "for order ID #", order.ID)
}
