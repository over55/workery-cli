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
	tag_ds "github.com/over55/workery-cli/app/tag/datastore"
	tenant_ds "github.com/over55/workery-cli/app/tenant/datastore"
	"github.com/over55/workery-cli/config"
)

func init() {
	rootCmd.AddCommand(importTagCmd)
}

var importTagCmd = &cobra.Command{
	Use:   "import_tag",
	Short: "Import the tags from old database",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		cfg := config.New()
		mc := mongodb.NewStorage(cfg)
		ppc := postgres.NewStorage(cfg, cfg.PostgresDB.DatabasePublicSchemaName)
		lpc := postgres.NewStorage(cfg, cfg.PostgresDB.DatabaseLondonSchemaName)

		defaultLogger := slog.Default()

		tenantStorer := tenant_ds.NewDatastore(cfg, defaultLogger, mc)
		irStorer := tag_ds.NewDatastore(cfg, defaultLogger, mc)

		tenant, err := tenantStorer.GetBySchemaName(ctx, cfg.PostgresDB.DatabaseLondonSchemaName)
		if err != nil {
			log.Fatal(err)
		}

		RunImportTag(cfg, ppc, lpc, irStorer, tenant)
	},
}

func RunImportTag(cfg *config.Conf, public *sql.DB, london *sql.DB, irStorer tag_ds.TagStorer, tenant *tenant_ds.Tenant) {
	fmt.Println("Beginning importing tags")
	data, err := ListAllTags(london)
	if err != nil {
		log.Fatal(err)
	}

	for _, datum := range data {
		importTag(context.Background(), irStorer, tenant, datum)
	}
	fmt.Println("Finished importing tags")
}

type OldUTag struct {
	ID          uint64 `json:"id"`
	Text        string `json:"text"`
	Description string `json:"description"`
	IsArchived  bool   `json:"is_archived"`
}

func ListAllTags(db *sql.DB) ([]*OldUTag, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `
	SELECT
	    id, text, description, is_archived
	FROM
	    workery_tags
	ORDER BY
	    id
	ASC
	`
	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}

	var arr []*OldUTag
	defer rows.Close()
	for rows.Next() {
		m := new(OldUTag)
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

func importTag(ctx context.Context, irStorer tag_ds.TagStorer, tenant *tenant_ds.Tenant, t *OldUTag) {
	var state int8 = 1
	if t.IsArchived == true {
		state = 2
	}

	m := &tag_ds.Tag{
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
	fmt.Println("Imported Tag ID#", m.ID)
}
