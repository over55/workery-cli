package cmd

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"log/slog"
	"time"

	"github.com/spf13/cobra"

	"github.com/over55/workery-cli/adapter/storage/mongodb"
	"github.com/over55/workery-cli/adapter/storage/postgres"
	a_ds "github.com/over55/workery-cli/app/order/datastore"
	ss_ds "github.com/over55/workery-cli/app/skillset/datastore"
	tenant_ds "github.com/over55/workery-cli/app/tenant/datastore"
	"github.com/over55/workery-cli/config"
)

func init() {
	rootCmd.AddCommand(importOrderSkillSetCmd)
}

var importOrderSkillSetCmd = &cobra.Command{
	Use:   "import_order_skill_set",
	Short: "Import the order skill sets from old database",
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
		oStorer := a_ds.NewDatastore(cfg, defaultLogger, mc)

		tenant, err := tenantStorer.GetBySchemaName(ctx, cfg.PostgresDB.DatabaseLondonSchemaName)
		if err != nil {
			log.Fatal(err)
		}

		RunImportOrderSkillSet(cfg, ppc, lpc, vtStorer, oStorer, tenant)
	},
}

func RunImportOrderSkillSet(cfg *config.Conf, public *sql.DB, london *sql.DB, irStorer ss_ds.SkillSetStorer, oStorer a_ds.OrderStorer, tenant *tenant_ds.Tenant) {
	fmt.Println("Beginning importing order skillsets")
	data, err := ListAllOrderSkillSets(london)
	if err != nil {
		log.Fatal(err)
	}

	for _, datum := range data {
		importOrderSkillSet(context.Background(), irStorer, oStorer, tenant, datum)
	}
	fmt.Println("Finished importing order skillsets")
}

type OldOrderSkillSet struct {
	ID         uint64 `json:"id"`
	OrderID    uint64 `json:"workorder_id"`
	SkillSetID uint64 `json:"skillset_id"`
}

func ListAllOrderSkillSets(db *sql.DB) ([]*OldOrderSkillSet, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `
	SELECT
        id, workorder_id, skillset_id
	FROM
        london.workery_work_orders_skill_sets
	ORDER BY
	    id
	ASC
	`
	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}

	var arr []*OldOrderSkillSet
	defer rows.Close()
	for rows.Next() {
		m := new(OldOrderSkillSet)
		err = rows.Scan(
			&m.ID,
			&m.OrderID,
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

func importOrderSkillSet(ctx context.Context, ssStorer ss_ds.SkillSetStorer, oStorer a_ds.OrderStorer, tenant *tenant_ds.Tenant, oa *OldOrderSkillSet) {
	//
	// Lookup related.
	//

	o, err := oStorer.GetByWJID(ctx, oa.OrderID)
	if err != nil {
		log.Fatal(err)
	}
	if o == nil {
		log.Fatal("order does not exist")
	}
	ss, err := ssStorer.GetByPublicID(ctx, oa.SkillSetID)
	if err != nil {
		log.Fatal(err)
	}
	if ss == nil {
		log.Fatal("skill set does not exist")
	}

	//
	// Create the order vehicle type.
	//

	avt := &a_ds.OrderSkillSet{
		ID:          ss.ID,
		Category:    ss.Category,
		SubCategory: ss.SubCategory,
		Description: ss.Description,
		Status:      ss.Status,
	}

	o.SkillSets = append(o.SkillSets, avt)

	if err := oStorer.UpdateByID(ctx, o); err != nil {
		log.Fatal(err)
	}
	fmt.Println("Imported order skill set ID#", avt.ID, "order ID #", o.ID)
}
