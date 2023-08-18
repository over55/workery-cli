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
	"gopkg.in/guregu/null.v4"

	"github.com/over55/workery-cli/adapter/storage/mongodb"
	"github.com/over55/workery-cli/adapter/storage/postgres"
	bulletin_ds "github.com/over55/workery-cli/app/bulletin/datastore"
	tenant_ds "github.com/over55/workery-cli/app/tenant/datastore"
	user_ds "github.com/over55/workery-cli/app/user/datastore"
	"github.com/over55/workery-cli/config"
)

func init() {
	rootCmd.AddCommand(importBulletinCmd)
}

var importBulletinCmd = &cobra.Command{
	Use:   "import_bulletins",
	Short: "Import the bulletins from old database",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		cfg := config.New()
		mc := mongodb.NewStorage(cfg)
		ppc := postgres.NewStorage(cfg, cfg.PostgresDB.DatabasePublicSchemaName)
		lpc := postgres.NewStorage(cfg, cfg.PostgresDB.DatabaseLondonSchemaName)

		defaultLogger := slog.Default()

		tenantStorer := tenant_ds.NewDatastore(cfg, defaultLogger, mc)
		cStorer := bulletin_ds.NewDatastore(cfg, defaultLogger, mc)
		userStorer := user_ds.NewDatastore(cfg, defaultLogger, mc)

		tenant, err := tenantStorer.GetBySchemaName(ctx, cfg.PostgresDB.DatabaseLondonSchemaName)
		if err != nil {
			log.Fatal(err)
		}

		RunImportBulletin(cfg, ppc, lpc, cStorer, userStorer, tenant)
	},
}

func RunImportBulletin(cfg *config.Conf, public *sql.DB, london *sql.DB, cStorer bulletin_ds.BulletinStorer, userStorer user_ds.UserStorer, tenant *tenant_ds.Tenant) {
	fmt.Println("Beginning importing bulletins")
	data, err := ListAllBulletinBoardItems(london)
	if err != nil {
		log.Fatal(err)
	}

	for _, datum := range data {
		importBulletin(context.Background(), cStorer, userStorer, tenant, datum)
	}
	fmt.Println("Finished importing bulletins")
}

type OldBulletinBoardItem struct {
	ID               uint64    `json:"id"`
	Text             string    `json:"text"`
	CreatedAt        time.Time `json:"created_at"`
	CreatedByID      null.Int  `json:"created_by_id,omitempty"`
	CreatedFrom      string    `json:"created_from"`
	LastModifiedAt   time.Time `json:"last_modified_at"`
	LastModifiedByID null.Int  `json:"last_modified_by_id,omitempty"`
	LastModifiedFrom string    `json:"last_modified_from"`
	IsArchived       bool      `json:"is_archived"`
}

func ListAllBulletinBoardItems(db *sql.DB) ([]*OldBulletinBoardItem, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `
	SELECT
	    id, text, created_at, created_by_id, created_from, last_modified_at, last_modified_by_id, last_modified_from, is_archived
	FROM
	    workery_bulletin_board_items
	ORDER BY
	    id
	ASC
	`
	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}

	var arr []*OldBulletinBoardItem
	defer rows.Close()
	for rows.Next() {
		m := new(OldBulletinBoardItem)
		err = rows.Scan(
			&m.ID,
			&m.Text,
			&m.CreatedAt,
			&m.CreatedByID,
			&m.CreatedFrom,
			&m.LastModifiedAt,
			&m.LastModifiedByID,
			&m.LastModifiedFrom,
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

func importBulletin(ctx context.Context, cStorer bulletin_ds.BulletinStorer, userStorer user_ds.UserStorer, tenant *tenant_ds.Tenant, oir *OldBulletinBoardItem) {
	//
	// Set the `state`.
	//

	var state int8 = 1
	if oir.IsArchived == true {
		state = 0
	}

	//
	// Get `createdByID` and `createdByName` values.
	//

	var createdByID primitive.ObjectID = primitive.NilObjectID
	var createdByName string
	if oir.CreatedByID.ValueOrZero() > 0 {
		user, err := userStorer.GetByOldID(ctx, uint64(oir.CreatedByID.ValueOrZero()))
		if err != nil {
			log.Fatal("ur.GetByOldID", err)
		}
		if user != nil {
			createdByID = user.ID
			createdByName = user.Name
		}

		// // For debugging purposes only.
		// log.Println("createdByID:", createdByID)
		// log.Println("createdByName:", createdByName)
	}

	//
	// Get `modifiedByID` and `modifiedByName` values.
	//

	var modifiedByID primitive.ObjectID = primitive.NilObjectID
	var modifiedByName string
	if oir.CreatedByID.ValueOrZero() > 0 {
		user, err := userStorer.GetByOldID(ctx, uint64(oir.CreatedByID.ValueOrZero()))
		if err != nil {
			log.Fatal("ur.GetByOldID", err)
		}
		if user != nil {
			modifiedByID = user.ID
			modifiedByName = user.Name
		}

		// // For debugging purposes only.
		// log.Println("modifiedByID:", modifiedByID)
		// log.Println("modifiedByName:", modifiedByName)
	}

	m := &bulletin_ds.Bulletin{
		ID:                    primitive.NewObjectID(),
		TenantID:              tenant.ID,
		CreatedAt:             oir.CreatedAt,
		CreatedByUserID:       createdByID,
		CreatedByUserName:     createdByName,
		CreatedFromIPAddress:  oir.CreatedFrom,
		ModifiedAt:            oir.LastModifiedAt,
		ModifiedByUserID:      modifiedByID,
		ModifiedByUserName:    modifiedByName,
		ModifiedFromIPAddress: oir.LastModifiedFrom,
		Text:                  oir.Text,
		Status:                state,
		OldID:                 oir.ID,
	}
	err := cStorer.Create(ctx, m)
	if err != nil {
		log.Fatal("ur.GetByOldID", err)
	}
	fmt.Println("Imported Bulletin ID#", m.ID)
}
