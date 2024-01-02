package cmd

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"log/slog"

	"github.com/spf13/cobra"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"gopkg.in/guregu/null.v4"

	"github.com/over55/workery-cli/adapter/storage/mongodb"
	"github.com/over55/workery-cli/adapter/storage/postgres"
	sf_ds "github.com/over55/workery-cli/app/servicefee/datastore"
	tenant_ds "github.com/over55/workery-cli/app/tenant/datastore"
	"github.com/over55/workery-cli/config"
)

func init() {
	rootCmd.AddCommand(importServiceFeeCmd)
}

var importServiceFeeCmd = &cobra.Command{
	Use:   "import_service_fee",
	Short: "Import the service fees from old database",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		cfg := config.New()
		mc := mongodb.NewStorage(cfg)
		ppc := postgres.NewStorage(cfg, cfg.PostgresDB.DatabasePublicSchemaName)
		lpc := postgres.NewStorage(cfg, cfg.PostgresDB.DatabaseLondonSchemaName)

		defaultLogger := slog.Default()

		tenantStorer := tenant_ds.NewDatastore(cfg, defaultLogger, mc)
		irStorer := sf_ds.NewDatastore(cfg, defaultLogger, mc)

		tenant, err := tenantStorer.GetBySchemaName(ctx, cfg.PostgresDB.DatabaseLondonSchemaName)
		if err != nil {
			log.Fatal(err)
		}

		RunImportServiceFee(cfg, ppc, lpc, irStorer, tenant)
	},
}

func RunImportServiceFee(cfg *config.Conf, public *sql.DB, london *sql.DB, irStorer sf_ds.ServiceFeeStorer, tenant *tenant_ds.Tenant) {
	fmt.Println("Beginning importing service fees")
	data, err := ListAllServiceFees(london)
	if err != nil {
		log.Fatal(err)
	}

	for _, datum := range data {
		importServiceFee(context.Background(), irStorer, tenant, datum)
	}
	fmt.Println("Finished importing service fees")
}

type OldUWorkOrderServiceFee struct {
	ID               uint64    `json:"id"`
	Title            string    `json:"title"`
	Description      string    `json:"description"`
	Percentage       float64   `json:"percentage"`
	CreatedAt        time.Time `json:"created_at"`
	CreatedByID      null.Int  `json:"created_by_id,omitempty"`
	LastModifiedAt   time.Time `json:"last_modified_at"`
	LastModifiedByID null.Int  `json:"last_modified_by_id,omitempty"`
	IsArchived       bool      `json:"is_archived"`
}

func ListAllServiceFees(db *sql.DB) ([]*OldUWorkOrderServiceFee, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `
	SELECT
	    id, title, description, percentage, created_at, created_by_id, last_modified_at, last_modified_by_id, is_archived
	FROM
	    workery_work_order_service_fees
	ORDER BY
	    id
	ASC
	`
	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}

	var arr []*OldUWorkOrderServiceFee
	defer rows.Close()
	for rows.Next() {
		m := new(OldUWorkOrderServiceFee)
		err = rows.Scan(
			&m.ID,
			&m.Title,
			&m.Description,
			&m.Percentage,
			&m.CreatedAt,
			&m.CreatedByID,
			&m.LastModifiedAt,
			&m.LastModifiedByID,
			&m.IsArchived,
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

func importServiceFee(ctx context.Context, irStorer sf_ds.ServiceFeeStorer, tenant *tenant_ds.Tenant, t *OldUWorkOrderServiceFee) {
	var state int8 = 1
	if t.IsArchived == true {
		state = 2
	}

	m := &sf_ds.ServiceFee{
		PublicID:       t.ID,
		ID:          primitive.NewObjectID(),
		Name:        t.Title,
		Percentage:  t.Percentage,
		Description: t.Description,
		Status:      state,
		TenantID:    tenant.ID,
	}
	err := irStorer.Create(ctx, m)
	if err != nil {
		log.Panic(err)
	}
	fmt.Println("Imported ServiceFee ID#", m.ID)
}
