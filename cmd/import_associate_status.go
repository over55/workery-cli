package cmd

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	"log/slog"

	"github.com/spf13/cobra"

	"github.com/over55/workery-cli/adapter/storage/mongodb"
	"github.com/over55/workery-cli/adapter/storage/postgres"
	a_ds "github.com/over55/workery-cli/app/associate/datastore"
	hh_ds "github.com/over55/workery-cli/app/howhear/datastore"
	sf_ds "github.com/over55/workery-cli/app/servicefee/datastore"
	tenant_ds "github.com/over55/workery-cli/app/tenant/datastore"
	user_ds "github.com/over55/workery-cli/app/user/datastore"
	"github.com/over55/workery-cli/config"
)

func init() {
	rootCmd.AddCommand(importAssociateStatusCmd)
}

var importAssociateStatusCmd = &cobra.Command{
	Use:   "import_associate_status",
	Short: "Adjust which associate is active based on hard coded values",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.New()
		mc := mongodb.NewStorage(cfg)
		ppc := postgres.NewStorage(cfg, cfg.PostgresDB.DatabasePublicSchemaName)
		lpc := postgres.NewStorage(cfg, cfg.PostgresDB.DatabasePublicSchemaName)
		defaultLogger := slog.Default()
		tenantStorer := tenant_ds.NewDatastore(cfg, defaultLogger, mc)
		userStorer := user_ds.NewDatastore(cfg, defaultLogger, mc)
		aStorer := a_ds.NewDatastore(cfg, defaultLogger, mc)
		hhStorer := hh_ds.NewDatastore(cfg, defaultLogger, mc)
		sfStorer := sf_ds.NewDatastore(cfg, defaultLogger, mc)

		tenant, err := tenantStorer.GetBySchemaName(context.Background(), cfg.PostgresDB.DatabaseLondonSchemaName)
		if err != nil {
			log.Fatal(err)
		}

		RunImportAssociateStatus(cfg, ppc, lpc, tenantStorer, userStorer, aStorer, hhStorer, sfStorer, tenant)
	},
}

// contains checks if a value is present in a slice.
func contains(s []uint64, val uint64) bool {
	for _, v := range s {
		if v == val {
			return true
		}
	}
	return false
}

func RunImportAssociateStatus(cfg *config.Conf, public *sql.DB, london *sql.DB, tenantStorer tenant_ds.TenantStorer, userStorer user_ds.UserStorer, aStorer a_ds.AssociateStorer, hhStorer hh_ds.HowHearAboutUsItemStorer, sfStorer sf_ds.ServiceFeeStorer, tenant *tenant_ds.Tenant) {
	fmt.Println("Beginning importing associate statuses")
	oaIDs := []uint64{
		6189,
		6187,
		6147,
		6023,
		6042,
		6164,
		6148,
		6171,
		6091,
		6144,
		6159,
		6116,
		6030,
		6185,
		6073,
		6169,
		6071,
		6186,
		6188,
		6146,
		6072,
		6130,
		5988,
		6097,
	}
	f := &a_ds.AssociatePaginationListFilter{
		Cursor:    "",
		PageSize:  1_000_000,
		SortField: "", // Forget sorting, we don't need it here.
		SortOrder: 1,
	}
	res, err := aStorer.ListByFilter(context.Background(), f)
	if err != nil {
		log.Fatal(err)
	}

	for _, a := range res.Results {
		// Check if a.OldID is in the oaIDs array
		if contains(oaIDs, a.OldID) {
			a.Status = a_ds.AssociateStatusActive
			fmt.Println("Changed Associate ID#", a.ID, "to status", a_ds.AssociateStatusActive)
		} else {
			a.Status = a_ds.AssociateStatusArchived
		}
		if err := aStorer.UpdateByID(context.Background(), a); err != nil {
			log.Fatal(err)
		}
	}
	fmt.Println("Finished importing associate statuses")
}
