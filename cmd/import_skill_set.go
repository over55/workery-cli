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
	ss_ds "github.com/over55/workery-cli/app/skillset/datastore"
	tenant_ds "github.com/over55/workery-cli/app/tenant/datastore"
	"github.com/over55/workery-cli/config"
)

func init() {
	rootCmd.AddCommand(importSkillSetCmd)
}

var importSkillSetCmd = &cobra.Command{
	Use:   "import_skill_set",
	Short: "Import skill sets from old database",
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

		tenant, err := tenantStorer.GetBySchemaName(ctx, cfg.PostgresDB.DatabaseLondonSchemaName)
		if err != nil {
			log.Fatal(err)
		}

		RunImportSkillSet(cfg, ppc, lpc, ssStorer, tenant)
	},
}

func RunImportSkillSet(cfg *config.Conf, public *sql.DB, london *sql.DB, ssStorer ss_ds.SkillSetStorer, tenant *tenant_ds.Tenant) {
	fmt.Println("Beginning importing skill sets")
	data, err := ListAllSkillSets(london)
	if err != nil {
		log.Fatal(err)
	}
	for _, datum := range data {
		importSkillSet(context.Background(), ssStorer, tenant, datum)
	}
	fmt.Println("Finished importing skill sets")
}

type OldUSkillSet struct {
	ID          uint64 `json:"id"`
	Category    string `json:"category"`
	SubCategory string `json:"sub_category"`
	Description string `json:"description"`
	IsArchived  bool   `json:"is_archived"`
	OldId       uint64 `json:"old_id"`
}

type SkillSet struct {
	ID          primitive.ObjectID `bson:"_id" json:"_id"`
	TenantID    primitive.ObjectID `bson:"tenant_id" json:"tenant_id"`
	TenantName  string             `bson:"tenant_name" json:"tenant_name"`
	Category    string             `bson:"category" json:"category"`
	SubCategory string             `bson:"sub_category" json:"sub_category"`
	Description string             `bson:"description" json:"description"`
	State       int8               `bson:"state" json:"state"`
	OldID       uint64             `bson:"old_id" json:"old_id"`
	// InsuranceRequirements []*SkillSetInsuranceRequirement `json:"skill_set_requirements,omitempty"` // Reference
}

// Function returns a paginated list of all type element items.
func ListAllSkillSets(db *sql.DB) ([]*OldUSkillSet, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `
	SELECT
        id, category, sub_category, description, is_archived
	FROM
        workery_skill_sets
	ORDER BY
	    id
	ASC
	`
	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		log.Println("failled querring old database")
		return nil, err
	}

	var arr []*OldUSkillSet
	defer rows.Close()
	for rows.Next() {
		m := new(OldUSkillSet)
		err = rows.Scan(
			&m.ID,
			&m.Category,
			&m.SubCategory,
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

func importSkillSet(ctx context.Context, ssStorer ss_ds.SkillSetStorer, tenant *tenant_ds.Tenant, t *OldUSkillSet) {
	var state int8 = 1
	if t.IsArchived == true {
		state = 2
	}

	m := &ss_ds.SkillSet{
		OldID:                 t.ID,
		TenantID:              tenant.ID,
		ID:                    primitive.NewObjectID(),
		Category:              t.Category,
		SubCategory:           t.SubCategory,
		Description:           t.Description,
		Status:                state,
		InsuranceRequirements: make([]*ss_ds.SkillSetInsuranceRequirement, 0),
	}
	if err := ssStorer.Create(ctx, m); err != nil {
		log.Panic(err)
	}
	fmt.Println("Imported skill set ID#", m.ID)
}
