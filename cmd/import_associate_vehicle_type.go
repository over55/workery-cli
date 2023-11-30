package cmd

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"log/slog"

	"github.com/spf13/cobra"

	"github.com/over55/workery-cli/adapter/storage/mongodb"
	"github.com/over55/workery-cli/adapter/storage/postgres"
	a_ds "github.com/over55/workery-cli/app/associate/datastore"
	tenant_ds "github.com/over55/workery-cli/app/tenant/datastore"
	vt_ds "github.com/over55/workery-cli/app/vehicletype/datastore"
	"github.com/over55/workery-cli/config"
)

func init() {
	rootCmd.AddCommand(importAssociateVehicleTypeCmd)
}

var importAssociateVehicleTypeCmd = &cobra.Command{
	Use:   "import_associate_vehicle_type",
	Short: "Import the associate vehicle types from old database",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		cfg := config.New()
		mc := mongodb.NewStorage(cfg)
		ppc := postgres.NewStorage(cfg, cfg.PostgresDB.DatabasePublicSchemaName)
		lpc := postgres.NewStorage(cfg, cfg.PostgresDB.DatabaseLondonSchemaName)

		defaultLogger := slog.Default()

		tenantStorer := tenant_ds.NewDatastore(cfg, defaultLogger, mc)
		vtStorer := vt_ds.NewDatastore(cfg, defaultLogger, mc)
		aStorer := a_ds.NewDatastore(cfg, defaultLogger, mc)

		tenant, err := tenantStorer.GetBySchemaName(ctx, cfg.PostgresDB.DatabaseLondonSchemaName)
		if err != nil {
			log.Fatal(err)
		}

		RunImportAssociateVehicleType(cfg, ppc, lpc, vtStorer, aStorer, tenant)
	},
}

func RunImportAssociateVehicleType(cfg *config.Conf, public *sql.DB, london *sql.DB, irStorer vt_ds.VehicleTypeStorer, aStorer a_ds.AssociateStorer, tenant *tenant_ds.Tenant) {
	fmt.Println("Beginning importing vehicle types")
	data, err := ListAllAssociateVehicleTypes(london)
	if err != nil {
		log.Fatal(err)
	}

	for _, datum := range data {
		importAssociateVehicleType(context.Background(), irStorer, aStorer, tenant, datum)
	}
	fmt.Println("Finished importing vehicle types")
}

type OldAssociateVehicleType struct {
	ID            uint64 `json:"id"`
	AssociateID   uint64 `json:"associate_id"`
	VehicleTypeID uint64 `json:"vehicletype_id"`
}

func ListAllAssociateVehicleTypes(db *sql.DB) ([]*OldAssociateVehicleType, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `
	SELECT
        id, associate_id, vehicletype_id
	FROM
        london.workery_associates_vehicle_types
	ORDER BY
	    id
	ASC
	`
	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}

	var arr []*OldAssociateVehicleType
	defer rows.Close()
	for rows.Next() {
		m := new(OldAssociateVehicleType)
		err = rows.Scan(
			&m.ID,
			&m.AssociateID,
			&m.VehicleTypeID,
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

func importAssociateVehicleType(ctx context.Context, vtStorer vt_ds.VehicleTypeStorer, aStorer a_ds.AssociateStorer, tenant *tenant_ds.Tenant, oa *OldAssociateVehicleType) {
	//
	// Lookup related.
	//

	a, err := aStorer.GetByOldID(ctx, oa.AssociateID)
	if err != nil {
		log.Fatal(err)
	}
	if a == nil {
		log.Fatal("associate does not exist")
	}
	vt, err := vtStorer.GetByOldID(ctx, oa.VehicleTypeID)
	if err != nil {
		log.Fatal(err)
	}
	if vt == nil {
		log.Fatal("vehicle type does not exist")
	}

	//
	// Create the associate vehicle type.
	//

	avt := &a_ds.AssociateVehicleType{
		ID:          vt.ID,
		Name:        vt.Name,
		Description: vt.Description,
		Status:      vt.Status,
	}

	a.VehicleTypes = append(a.VehicleTypes, avt)

	if err := aStorer.UpdateByID(ctx, a); err != nil {
		log.Fatal(err)
	}
	fmt.Println("Imported associate vehicle type ID#", vt.ID, "associate ID#", a.ID)
}
