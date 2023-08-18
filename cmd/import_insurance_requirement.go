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

	"github.com/over55/workery-cli/adapter/storage/mongodb"
	"github.com/over55/workery-cli/adapter/storage/postgres"
	ir_ds "github.com/over55/workery-cli/app/insurancerequirement/datastore"
	tenant_ds "github.com/over55/workery-cli/app/tenant/datastore"
	"github.com/over55/workery-cli/config"
)

func init() {
	rootCmd.AddCommand(importInsuranceRequirementCmd)
}

var importInsuranceRequirementCmd = &cobra.Command{
	Use:   "import_insurance_requirement",
	Short: "Import the insurance requirement from old database",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		cfg := config.New()
		mc := mongodb.NewStorage(cfg)
		ppc := postgres.NewStorage(cfg, cfg.PostgresDB.DatabasePublicSchemaName)
		lpc := postgres.NewStorage(cfg, cfg.PostgresDB.DatabaseLondonSchemaName)

		defaultLogger := slog.Default()

		tenantStorer := tenant_ds.NewDatastore(cfg, defaultLogger, mc)
		irStorer := ir_ds.NewDatastore(cfg, defaultLogger, mc)

		tenant, err := tenantStorer.GetBySchemaName(ctx, cfg.PostgresDB.DatabaseLondonSchemaName)
		if err != nil {
			log.Fatal(err)
		}

		RunImportInsuranceRequirement(cfg, ppc, lpc, irStorer, tenant)
	},
}

func RunImportInsuranceRequirement(cfg *config.Conf, public *sql.DB, london *sql.DB, irStorer ir_ds.InsuranceRequirementStorer, tenant *tenant_ds.Tenant) {
	fmt.Println("Beginning importing insurance requirements")
	data, err := ListAllInsuranceRequirements(london)
	if err != nil {
		log.Fatal(err)
	}

	for _, datum := range data {
		importInsuranceRequirement(context.Background(), irStorer, tenant, datum)
	}
	fmt.Println("Finished importing insurance requirements")
}

type OldUInsuranceRequirement struct {
	ID          uint64 `json:"id"`
	Text        string `json:"text"`
	Description string `json:"description"`
	IsArchived  bool   `json:"is_archived"`
}

// Function returns a paginated list of all type element items.
func ListAllInsuranceRequirements(db *sql.DB) ([]*OldUInsuranceRequirement, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `
    SELECT
	    id, text, description, is_archived
	FROM
	    workery_insurance_requirements
	ORDER BY
	    id
	ASC
	`
	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		log.Println("failled querring old database")
		return nil, err
	}

	var arr []*OldUInsuranceRequirement
	defer rows.Close()
	for rows.Next() {
		m := new(OldUInsuranceRequirement)
		err = rows.Scan(
			&m.ID,
			&m.Text,
			&m.Description,
			&m.IsArchived,
		)
		if err != nil {
			log.Println("failled querring2 old database")
			panic(err)
		}
		arr = append(arr, m)
	}
	err = rows.Err()
	if err != nil {
		log.Println("failled querring3 old database")
		panic(err)
	}
	return arr, err
}

func importInsuranceRequirement(ctx context.Context, irStorer ir_ds.InsuranceRequirementStorer, tenant *tenant_ds.Tenant, t *OldUInsuranceRequirement) {
	var state int8 = 1
	if t.IsArchived == true {
		state = 2
	}

	m := &ir_ds.InsuranceRequirement{
		OldID:       t.ID,
		ID:          primitive.NewObjectID(),
		Text:        t.Text,
		Description: t.Description,
		Status:      state,
		TenantID:    tenant.ID,
	}
	err := irStorer.Create(ctx, m)
	if err != nil {
		log.Panic(err)
	}
	fmt.Println("Imported Insurance requirement ID#", m.ID)
}
