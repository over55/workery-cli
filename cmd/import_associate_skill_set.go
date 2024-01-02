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
	ss_ds "github.com/over55/workery-cli/app/skillset/datastore"
	tenant_ds "github.com/over55/workery-cli/app/tenant/datastore"
	"github.com/over55/workery-cli/config"
)

func init() {
	rootCmd.AddCommand(importAssociateSkillSetCmd)
}

var importAssociateSkillSetCmd = &cobra.Command{
	Use:   "import_associate_skill_set",
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
		vtStorer := ss_ds.NewDatastore(cfg, defaultLogger, mc)
		aStorer := a_ds.NewDatastore(cfg, defaultLogger, mc)

		tenant, err := tenantStorer.GetBySchemaName(ctx, cfg.PostgresDB.DatabaseLondonSchemaName)
		if err != nil {
			log.Fatal(err)
		}

		RunImportAssociateSkillSet(cfg, ppc, lpc, vtStorer, aStorer, tenant)
	},
}

func RunImportAssociateSkillSet(cfg *config.Conf, public *sql.DB, london *sql.DB, irStorer ss_ds.SkillSetStorer, aStorer a_ds.AssociateStorer, tenant *tenant_ds.Tenant) {
	fmt.Println("Beginning importing associate skillsets")
	data, err := ListAllAssociateSkillSets(london)
	if err != nil {
		log.Fatal(err)
	}

	for _, datum := range data {
		importAssociateSkillSet(context.Background(), irStorer, aStorer, tenant, datum)
	}
	fmt.Println("Finished importing associate skillsets")
}

type OldAssociateSkillSet struct {
	ID          uint64 `json:"id"`
	AssociateID uint64 `json:"associate_id"`
	SkillSetID  uint64 `json:"skillset_id"`
}

func ListAllAssociateSkillSets(db *sql.DB) ([]*OldAssociateSkillSet, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `
	SELECT
        id, associate_id, skillset_id
	FROM
        london.workery_associates_skill_sets
	ORDER BY
	    id
	ASC
	`
	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}

	var arr []*OldAssociateSkillSet
	defer rows.Close()
	for rows.Next() {
		m := new(OldAssociateSkillSet)
		err = rows.Scan(
			&m.ID,
			&m.AssociateID,
			&m.SkillSetID,
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

func importAssociateSkillSet(ctx context.Context, ssStorer ss_ds.SkillSetStorer, aStorer a_ds.AssociateStorer, tenant *tenant_ds.Tenant, oa *OldAssociateSkillSet) {
	//
	// Lookup related.
	//

	a, err := aStorer.GetByPublicID(ctx, oa.AssociateID)
	if err != nil {
		log.Fatal(err)
	}
	if a == nil {
		log.Fatal("associate does not exist")
	}
	ss, err := ssStorer.GetByPublicID(ctx, oa.SkillSetID)
	if err != nil {
		log.Fatal(err)
	}
	if ss == nil {
		log.Fatal("skill set does not exist")
	}

	//
	// Create the associate vehicle type.
	//

	avt := &a_ds.AssociateSkillSet{
		ID:          ss.ID,
		Category:    ss.Category,
		SubCategory: ss.SubCategory,
		Description: ss.Description,
		Status:      ss.Status,
	}

	a.SkillSets = append(a.SkillSets, avt)

	if err := aStorer.UpdateByID(ctx, a); err != nil {
		log.Fatal(err)
	}
	fmt.Println("Imported associate skill set ID#", avt.ID, "associate ID #", a.ID)
}
