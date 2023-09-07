package cmd

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/spf13/cobra"
	"golang.org/x/exp/slog"

	"github.com/over55/workery-cli/adapter/storage/mongodb"
	"github.com/over55/workery-cli/adapter/storage/postgres"
	ir_ds "github.com/over55/workery-cli/app/insurancerequirement/datastore"
	ss_ds "github.com/over55/workery-cli/app/skillset/datastore"
	tenant_ds "github.com/over55/workery-cli/app/tenant/datastore"
	"github.com/over55/workery-cli/config"
)

func init() {
	rootCmd.AddCommand(importSkillSetInsuranceRequirementCmd)
}

var importSkillSetInsuranceRequirementCmd = &cobra.Command{
	Use:   "import_skill_set_insurance_requirement",
	Short: "Import skill set insurance requirements from old database",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		cfg := config.New()
		mc := mongodb.NewStorage(cfg)
		ppc := postgres.NewStorage(cfg, cfg.PostgresDB.DatabasePublicSchemaName)
		lpc := postgres.NewStorage(cfg, cfg.PostgresDB.DatabaseLondonSchemaName)

		defaultLogger := slog.Default()

		tenantStorer := tenant_ds.NewDatastore(cfg, defaultLogger, mc)
		ssStorer := ss_ds.NewDatastore(cfg, defaultLogger, mc)
		irStorer := ir_ds.NewDatastore(cfg, defaultLogger, mc)

		tenant, err := tenantStorer.GetBySchemaName(ctx, cfg.PostgresDB.DatabaseLondonSchemaName)
		if err != nil {
			log.Fatal(err)
		}

		RunImportSkillSetInsuranceRequirement(cfg, ppc, lpc, ssStorer, irStorer, tenant)
	},
}

type OldSkillSetInsuranceRequirement struct {
	Id                     uint64 `json:"id"`
	SkillSetId             uint64 `json:"skill_set_id"`
	InsuranceRequirementId uint64 `json:"insurance_requirement_id"`
}

func ListAllSkillSetInsuranceRequirements(db *sql.DB) ([]*OldSkillSetInsuranceRequirement, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `
	SELECT
        id, skillset_id, insurancerequirement_id
	FROM
        workery_skill_sets_insurance_requirements
	ORDER BY
	    id
	ASC
	`
	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}

	var arr []*OldSkillSetInsuranceRequirement
	defer rows.Close()
	for rows.Next() {
		m := new(OldSkillSetInsuranceRequirement)
		err = rows.Scan(
			&m.Id,
			&m.SkillSetId,
			&m.InsuranceRequirementId,
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

func RunImportSkillSetInsuranceRequirement(cfg *config.Conf, public *sql.DB, london *sql.DB, ssStorer ss_ds.SkillSetStorer, irStorer ir_ds.InsuranceRequirementStorer, tenant *tenant_ds.Tenant) {
	fmt.Println("Beginning importing skill sets")
	data, err := ListAllSkillSetInsuranceRequirements(london)
	if err != nil {
		log.Fatal(err)
	}
	for _, datum := range data {
		importSkillSetInsuranceRequirement(context.Background(), ssStorer, irStorer, tenant, datum)
	}
	fmt.Println("Finished importing skill sets")
}

func importSkillSetInsuranceRequirement(ctx context.Context, ssStorer ss_ds.SkillSetStorer, irStorer ir_ds.InsuranceRequirementStorer, tenant *tenant_ds.Tenant, t *OldSkillSetInsuranceRequirement) {
	ss, err := ssStorer.GetByOldID(ctx, t.SkillSetId)
	if err != nil {
		log.Fatal(err)
	}
	if ss == nil {
		log.Fatal("ss does not exist")
	}
	ir, err := irStorer.GetByOldID(ctx, t.InsuranceRequirementId)
	if err != nil {
		log.Fatal(err)
	}
	if ir == nil {
		log.Fatal("ss does not exist")
	}

	m := &ss_ds.SkillSetInsuranceRequirement{
		SkillSetID:  ss.ID,
		TenantID:    tenant.ID,
		ID:          ir.ID,
		Name:        ir.Name,
		Description: ir.Description,
		Status:      1, // 1=Active
		OldID:       ir.OldID,
	}
	ss.InsuranceRequirements = append(ss.InsuranceRequirements, m)

	if err := ssStorer.UpdateByID(ctx, ss); err != nil {
		log.Fatal(err)
	}
	fmt.Println("Imported insurance requirement for skill set ID#", ss.ID)
}
