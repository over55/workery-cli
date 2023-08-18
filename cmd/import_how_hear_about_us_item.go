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
	hh_ds "github.com/over55/workery-cli/app/howhear/datastore"
	tenant_ds "github.com/over55/workery-cli/app/tenant/datastore"
	"github.com/over55/workery-cli/config"
)

func init() {
	rootCmd.AddCommand(importHowHearAboutUsItemCmd)
}

var importHowHearAboutUsItemCmd = &cobra.Command{
	Use:   "import_how_hear_about_us_item",
	Short: "Import the how hear about us item from old database",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		cfg := config.New()
		mc := mongodb.NewStorage(cfg)
		ppc := postgres.NewStorage(cfg, cfg.PostgresDB.DatabasePublicSchemaName)
		lpc := postgres.NewStorage(cfg, cfg.PostgresDB.DatabaseLondonSchemaName)

		defaultLogger := slog.Default()

		tenantStorer := tenant_ds.NewDatastore(cfg, defaultLogger, mc)
		hhStorer := hh_ds.NewDatastore(cfg, defaultLogger, mc)

		tenant, err := tenantStorer.GetBySchemaName(ctx, cfg.PostgresDB.DatabaseLondonSchemaName)
		if err != nil {
			log.Fatal(err)
		}

		RunImportHowHearAboutUsItem(cfg, ppc, lpc, hhStorer, tenant)
	},
}

func RunImportHowHearAboutUsItem(cfg *config.Conf, public *sql.DB, london *sql.DB, hhStorer hh_ds.HowHearAboutUsItemStorer, tenant *tenant_ds.Tenant) {
	fmt.Println("Beginning importing how hear about us item")
	data, err := ListAllHowHearAboutUsItems(london)
	if err != nil {
		log.Fatal(err)
	}

	for _, datum := range data {
		importHowHearAboutUsItem(context.Background(), hhStorer, tenant, datum)
	}
	fmt.Println("Finished importing how hear about us item")
}

type OldUHowHearAboutUsItem struct {
	ID             uint64 `json:"id"`
	TenantID       uint64 `json:"tenant_id"`
	Text           string `json:"text"`
	SortNumber     int8   `json:"sort_number"`
	IsForAssociate bool   `json:"is_for_associate"`
	IsForCustomer  bool   `json:"is_for_customer"`
	IsForStaff     bool   `json:"is_for_staff"`
	IsArchived     bool   `json:"is_archived"`
}

type HowHearAboutUsItem struct {
	ID             primitive.ObjectID `bson:"_id" json:"id"`
	TenantID       primitive.ObjectID `bson:"tenant_id" json:"tenant_id"`
	TenantName     string             `bson:"tenant_name" json:"tenant_name"`
	Text           string             `bson:"text" json:"text"`
	SortNumber     int8               `bson:"sort_number" json:"sort_number"`
	IsForAssociate bool               `bson:"is_for_associate" json:"is_for_associate"`
	IsForCustomer  bool               `bson:"is_for_customer" json:"is_for_customer"`
	IsForStaff     bool               `bson:"is_for_staff" json:"is_for_staff"`
	State          int8               `bson:"state" json:"state"`
	OldID          uint64             `bson:"old_id" json:"old_id"`
}

// Function returns a paginated list of all type element items.
func ListAllHowHearAboutUsItems(db *sql.DB) ([]*OldUHowHearAboutUsItem, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `
	SELECT
        id, text, sort_number, is_for_associate, is_for_customer,
		is_for_staff, is_archived
	FROM
        workery_how_hear_about_us_items
	ORDER BY
	    id
	ASC
	`
	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		log.Println("failled querring old database")
		return nil, err
	}

	var arr []*OldUHowHearAboutUsItem
	defer rows.Close()
	for rows.Next() {
		m := new(OldUHowHearAboutUsItem)
		err = rows.Scan(
			&m.ID, &m.Text, &m.SortNumber, &m.IsForAssociate, &m.IsForCustomer,
			&m.IsForStaff, &m.IsArchived,
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

func importHowHearAboutUsItem(ctx context.Context, hhStorer hh_ds.HowHearAboutUsItemStorer, tenant *tenant_ds.Tenant, t *OldUHowHearAboutUsItem) {
	var state int8 = 1
	if t.IsArchived == true {
		state = 2
	}

	m := &hh_ds.HowHearAboutUsItem{
		ID:             primitive.NewObjectID(),
		OldID:          t.ID,
		TenantID:       tenant.ID,
		Text:           t.Text,
		IsForAssociate: t.IsForAssociate,
		IsForCustomer:  t.IsForCustomer,
		IsForStaff:     t.IsForStaff,
		SortNumber:     1,
		Status:         state,
	}
	err := hhStorer.Create(ctx, m)
	if err != nil {
		log.Panic(err)
	}
	fmt.Println("Imported how hear about us item ID#", m.ID)
}
