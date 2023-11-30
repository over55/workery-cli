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
	"gopkg.in/guregu/null.v4"

	"github.com/over55/workery-cli/adapter/storage/mongodb"
	"github.com/over55/workery-cli/adapter/storage/postgres"
	comment_ds "github.com/over55/workery-cli/app/comment/datastore"
	tenant_ds "github.com/over55/workery-cli/app/tenant/datastore"
	user_ds "github.com/over55/workery-cli/app/user/datastore"
	"github.com/over55/workery-cli/config"
)

func init() {
	rootCmd.AddCommand(importCommentCmd)
}

var importCommentCmd = &cobra.Command{
	Use:   "import_comment",
	Short: "Import the comments from old database",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		cfg := config.New()
		mc := mongodb.NewStorage(cfg)
		ppc := postgres.NewStorage(cfg, cfg.PostgresDB.DatabasePublicSchemaName)
		lpc := postgres.NewStorage(cfg, cfg.PostgresDB.DatabaseLondonSchemaName)

		defaultLogger := slog.Default()

		tenantStorer := tenant_ds.NewDatastore(cfg, defaultLogger, mc)
		cStorer := comment_ds.NewDatastore(cfg, defaultLogger, mc)
		userStorer := user_ds.NewDatastore(cfg, defaultLogger, mc)

		tenant, err := tenantStorer.GetBySchemaName(ctx, cfg.PostgresDB.DatabaseLondonSchemaName)
		if err != nil {
			log.Fatal(err)
		}

		RunImportComment(cfg, ppc, lpc, cStorer, userStorer, tenant)
	},
}

func RunImportComment(cfg *config.Conf, public *sql.DB, london *sql.DB, cStorer comment_ds.CommentStorer, userStorer user_ds.UserStorer, tenant *tenant_ds.Tenant) {
	fmt.Println("Beginning importing comments")
	data, err := ListAllComments(london)
	if err != nil {
		log.Fatal(err)
	}

	for _, datum := range data {
		importComment(context.Background(), cStorer, userStorer, tenant, datum)
	}
	fmt.Println("Finished importing comments")
}

type OldComment struct {
	ID               uint64      `json:"id"`
	CreatedAt        time.Time   `json:"created_time"`
	CreatedByID      null.Int    `json:"created_by_id,omitempty"`
	CreatedFrom      null.String `json:"created_from"`
	LastModifiedAt   time.Time   `json:"last_modified_time"`
	LastModifiedByID null.Int    `json:"last_modified_by_id,omitempty"`
	LastModifiedFrom null.String `json:"last_modified_from"`
	Text             string      `json:"text"`
	IsArchived       bool        `json:"is_archived"`
}

func ListAllComments(db *sql.DB) ([]*OldComment, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `
	SELECT
		id, created_at, created_by_id, created_from, last_modified_at, last_modified_by_id, last_modified_from, text, is_archived
	FROM
	    workery_comments
	ORDER BY
	    id
	ASC
	`
	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}

	var arr []*OldComment
	defer rows.Close()
	for rows.Next() {
		m := new(OldComment)
		err = rows.Scan(
			&m.ID,
			&m.CreatedAt,
			&m.CreatedByID,
			&m.CreatedFrom,
			&m.LastModifiedAt,
			&m.LastModifiedByID,
			&m.LastModifiedFrom,
			&m.Text,
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

func importComment(ctx context.Context, cStorer comment_ds.CommentStorer, userStorer user_ds.UserStorer, tenant *tenant_ds.Tenant, oir *OldComment) {
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

	//
	// Insert `comment` record.
	//

	m := &comment_ds.Comment{
		ID:                    primitive.NewObjectID(),
		TenantID:              tenant.ID,
		CreatedAt:             oir.CreatedAt,
		CreatedByUserID:       createdByID,
		CreatedByUserName:     createdByName,
		CreatedFromIPAddress:  oir.CreatedFrom.ValueOrZero(),
		ModifiedAt:            oir.LastModifiedAt,
		ModifiedByUserID:      modifiedByID,
		ModifiedByUserName:    modifiedByName,
		ModifiedFromIPAddress: oir.LastModifiedFrom.ValueOrZero(),
		Content:               oir.Text,
		Status:                state,
		OldID:                 oir.ID,
	}
	err := cStorer.Create(ctx, m)
	if err != nil {
		log.Fatal("ur.GetByOldID", err)
	}
	fmt.Println("Imported Comment ID#", m.ID)
}
