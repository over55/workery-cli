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
	asso_ds "github.com/over55/workery-cli/app/associate/datastore"
	hh_ds "github.com/over55/workery-cli/app/howhear/datastore"
	ir_ds "github.com/over55/workery-cli/app/insurancerequirement/datastore"
	tenant_ds "github.com/over55/workery-cli/app/tenant/datastore"
	user_ds "github.com/over55/workery-cli/app/user/datastore"
	"github.com/over55/workery-cli/config"
)

func init() {
	rootCmd.AddCommand(importAssociateInsuranceRequirementCmd)
}

var importAssociateInsuranceRequirementCmd = &cobra.Command{
	Use:   "import_associate_insurance_requirement",
	Short: "Import the associate insurance requirement from old database",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.New()
		mc := mongodb.NewStorage(cfg)
		ppc := postgres.NewStorage(cfg, cfg.PostgresDB.DatabasePublicSchemaName)
		lpc := postgres.NewStorage(cfg, cfg.PostgresDB.DatabasePublicSchemaName)
		defaultLogger := slog.Default()
		tenantStorer := tenant_ds.NewDatastore(cfg, defaultLogger, mc)
		userStorer := user_ds.NewDatastore(cfg, defaultLogger, mc)
		aStorer := asso_ds.NewDatastore(cfg, defaultLogger, mc)
		hhStorer := hh_ds.NewDatastore(cfg, defaultLogger, mc)
		irStorer := ir_ds.NewDatastore(cfg, defaultLogger, mc)

		tenant, err := tenantStorer.GetBySchemaName(context.Background(), cfg.PostgresDB.DatabaseLondonSchemaName)
		if err != nil {
			log.Fatal(err)
		}

		RunImportAssociateInsuranceRequirement(cfg, ppc, lpc, tenantStorer, userStorer, aStorer, hhStorer, irStorer, tenant)
	},
}

type OldAssociateInsuranceRequirement struct {
	Id                     uint64 `json:"id"`
	AssociateId            uint64 `json:"associate_id"`
	InsuranceRequirementId uint64 `json:"insurancerequirement_id"`
}

func ListAllAssociateInsuranceRequirements(db *sql.DB) ([]*OldAssociateInsuranceRequirement, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `
	SELECT
        id, associate_id, insurancerequirement_id
	FROM
        london.workery_associates_insurance_requirements
	ORDER BY
	    id
	ASC
	`
	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}

	var arr []*OldAssociateInsuranceRequirement
	defer rows.Close()
	for rows.Next() {
		m := new(OldAssociateInsuranceRequirement)
		err = rows.Scan(
			&m.Id,
			&m.AssociateId,
			&m.InsuranceRequirementId,
		)
		if err != nil {
			log.Fatal("rows.Scan", err)
		}
		arr = append(arr, m)
	}
	err = rows.Err()
	if err != nil {
		log.Fatal("rows.Err", err)
	}
	return arr, err
}

func RunImportAssociateInsuranceRequirement(cfg *config.Conf, public *sql.DB, london *sql.DB, tenantStorer tenant_ds.TenantStorer, userStorer user_ds.UserStorer, aStorer asso_ds.AssociateStorer, hhStorer hh_ds.HowHearAboutUsItemStorer, irStorer ir_ds.InsuranceRequirementStorer, tenant *tenant_ds.Tenant) {
	fmt.Println("Beginning importing associates")
	data, err := ListAllAssociateInsuranceRequirements(london)
	if err != nil {
		log.Fatal(err)
	}

	for _, datum := range data {
		importAssociateInsuranceRequirement(context.Background(), tenantStorer, userStorer, aStorer, hhStorer, irStorer, tenant, datum)
	}
	fmt.Println("Finished importing associates")
}

func importAssociateInsuranceRequirement(ctx context.Context, ts tenant_ds.TenantStorer, us user_ds.UserStorer, aStorer asso_ds.AssociateStorer, hhStorer hh_ds.HowHearAboutUsItemStorer, irStorer ir_ds.InsuranceRequirementStorer, tenant *tenant_ds.Tenant, ou *OldAssociateInsuranceRequirement) {
	//
	// Lookup related.
	//

	associate, err := aStorer.GetByPublicID(ctx, ou.AssociateId)
	if err != nil {
		log.Fatal(err)
	}
	if associate == nil {
		log.Fatal("associate does not exist")
	}
	ir, err := irStorer.GetByPublicID(ctx, ou.InsuranceRequirementId)
	if err != nil {
		log.Fatal(err)
	}
	if ir == nil {
		log.Fatal("insurance requirement does not exist")
	}

	//
	// Create the associate comment.
	//

	air := &asso_ds.AssociateInsuranceRequirement{
		ID:          ir.ID,
		Name:        ir.Name,
		Description: ir.Description,
		Status:      ir.Status,
	}

	// Append comments to associate details.
	associate.InsuranceRequirements = append(associate.InsuranceRequirements, air)

	if err := aStorer.UpdateByID(ctx, associate); err != nil {
		log.Fatal(err)
	}

	fmt.Println("Imported Associate Insurance Requirement ID#", air.ID, "for AssociateID", associate.ID)
}
