package cmd

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/spf13/cobra"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"log/slog"

	"github.com/over55/workery-cli/adapter/storage/mongodb"
	"github.com/over55/workery-cli/adapter/storage/postgres"
	tenant_ds "github.com/over55/workery-cli/app/tenant/datastore"
	vt_ds "github.com/over55/workery-cli/app/vehicletype/datastore"
	"github.com/over55/workery-cli/config"
)

func init() {
	rootCmd.AddCommand(importVehicleTypeCmd)
}

var importVehicleTypeCmd = &cobra.Command{
	Use:   "import_vehicle_type",
	Short: "Import the vehicle types from old database",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		cfg := config.New()
		mc := mongodb.NewStorage(cfg)
		ppc := postgres.NewStorage(cfg, cfg.PostgresDB.DatabasePublicSchemaName)
		lpc := postgres.NewStorage(cfg, cfg.PostgresDB.DatabaseLondonSchemaName)

		defaultLogger := slog.Default()

		tenantStorer := tenant_ds.NewDatastore(cfg, defaultLogger, mc)
		irStorer := vt_ds.NewDatastore(cfg, defaultLogger, mc)

		tenant, err := tenantStorer.GetBySchemaName(ctx, cfg.PostgresDB.DatabaseLondonSchemaName)
		if err != nil {
			log.Fatal(err)
		}

		RunImportVehicleType(cfg, ppc, lpc, irStorer, tenant)
	},
}

func RunImportVehicleType(cfg *config.Conf, public *sql.DB, london *sql.DB, irStorer vt_ds.VehicleTypeStorer, tenant *tenant_ds.Tenant) {
	fmt.Println("Beginning importing vehicle types")
	data, err := ListAllVehicleTypes(london)
	if err != nil {
		log.Fatal(err)
	}

	for _, datum := range data {
		importVehicleType(context.Background(), irStorer, tenant, datum)
	}
	fmt.Println("Finished importing vehicle types")
}

type OldUVehicleType struct {
	ID          uint64 `json:"id"`
	Text        string `json:"text"`
	Description string `json:"description"`
	IsArchived  bool   `json:"is_archived"`
}

func ListAllVehicleTypes(db *sql.DB) ([]*OldUVehicleType, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `
	SELECT
	    id, text, description, is_archived
	FROM
	    workery_vehicle_types
	ORDER BY
	    id
	ASC
	`
	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}

	var arr []*OldUVehicleType
	defer rows.Close()
	for rows.Next() {
		m := new(OldUVehicleType)
		err = rows.Scan(
			&m.ID,
			&m.Text,
			&m.Description,
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

func importVehicleType(ctx context.Context, irStorer vt_ds.VehicleTypeStorer, tenant *tenant_ds.Tenant, t *OldUVehicleType) {
	var state int8 = 1
	if t.IsArchived == true {
		state = 2
	}

	m := &vt_ds.VehicleType{
		OldID:       t.ID,
		ID:          primitive.NewObjectID(),
		Name:        t.Text,
		Description: t.Description,
		Status:      state,
		TenantID:    tenant.ID,
	}
	err := irStorer.Create(ctx, m)
	if err != nil {
		log.Panic(err)
	}
	fmt.Println("Imported VehicleType ID#", m.ID)
}
