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
	asso_ds "github.com/over55/workery-cli/app/associate/datastore"
	comm_ds "github.com/over55/workery-cli/app/comment/datastore"
	hh_ds "github.com/over55/workery-cli/app/howhear/datastore"
	tenant_ds "github.com/over55/workery-cli/app/tenant/datastore"
	user_ds "github.com/over55/workery-cli/app/user/datastore"
	"github.com/over55/workery-cli/config"
)

func init() {
	rootCmd.AddCommand(importAssociateCommentCmd)
}

var importAssociateCommentCmd = &cobra.Command{
	Use:   "import_associate_comment",
	Short: "Import the associate comments from old database",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.New()
		mc := mongodb.NewStorage(cfg)
		ppc := postgres.NewStorage(cfg, cfg.PostgresDB.DatabasePublicSchemaName)
		lpc := postgres.NewStorage(cfg, cfg.PostgresDB.DatabasePublicSchemaName)
		defaultLogger := slog.Default()
		tenantStorer := tenant_ds.NewDatastore(cfg, defaultLogger, mc)
		userStorer := user_ds.NewDatastore(cfg, defaultLogger, mc)
		custStorer := asso_ds.NewDatastore(cfg, defaultLogger, mc)
		hhStorer := hh_ds.NewDatastore(cfg, defaultLogger, mc)
		comStorer := comm_ds.NewDatastore(cfg, defaultLogger, mc)

		tenant, err := tenantStorer.GetBySchemaName(context.Background(), cfg.PostgresDB.DatabaseLondonSchemaName)
		if err != nil {
			log.Fatal(err)
		}

		RunImportAssociateComment(cfg, ppc, lpc, tenantStorer, userStorer, custStorer, hhStorer, comStorer, tenant)
	},
}

type OldAssociateComment struct {
	Id          uint64    `json:"id"`
	CreatedAt   time.Time `json:"created_at"`
	AssociateId uint64    `json:"associate_id"`
	CommentId   uint64    `json:"comment_id"`
}

func ListAllAssociateComments(db *sql.DB) ([]*OldAssociateComment, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `
	SELECT
	    id, created_at, about_id, comment_id
	FROM
	    london.workery_associate_comments
	ORDER BY
	    id
	ASC
	`
	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}

	var arr []*OldAssociateComment
	defer rows.Close()
	for rows.Next() {
		m := new(OldAssociateComment)
		err = rows.Scan(
			&m.Id,
			&m.CreatedAt,
			&m.AssociateId,
			&m.CommentId,
		)
		if err != nil {
			log.Panic("ListAllAssociateComments | Next | err", err)
		}
		arr = append(arr, m)
	}
	err = rows.Err()
	if err != nil {
		panic(err)
	}
	return arr, err
}

func RunImportAssociateComment(cfg *config.Conf, public *sql.DB, london *sql.DB, tenantStorer tenant_ds.TenantStorer, userStorer user_ds.UserStorer, custStorer asso_ds.AssociateStorer, hhStorer hh_ds.HowHearAboutUsItemStorer, comStorer comm_ds.CommentStorer, tenant *tenant_ds.Tenant) {
	fmt.Println("Beginning importing associates")
	data, err := ListAllAssociateComments(london)
	if err != nil {
		log.Fatal(err)
	}

	for _, datum := range data {
		importAssociateComment(context.Background(), tenantStorer, userStorer, custStorer, hhStorer, comStorer, tenant, datum)
	}
	fmt.Println("Finished importing associates")
}

func importAssociateComment(ctx context.Context, ts tenant_ds.TenantStorer, us user_ds.UserStorer, custStorer asso_ds.AssociateStorer, hhStorer hh_ds.HowHearAboutUsItemStorer, comStorer comm_ds.CommentStorer, tenant *tenant_ds.Tenant, ou *OldAssociateComment) {

	//
	// Lookup related.
	//

	associate, err := custStorer.GetByOldID(ctx, ou.AssociateId)
	if err != nil {
		log.Fatal(err)
	}
	if associate == nil {
		log.Fatal("associate does not exist")
	}
	comment, err := comStorer.GetByOldID(ctx, ou.CommentId)
	if err != nil {
		log.Fatal(err)
	}
	if comment == nil {
		log.Fatal("comment does not exist")
	}

	//
	// Create the associate comment.
	//

	cc := &asso_ds.AssociateComment{
		ID:                    comment.ID,
		TenantID:              comment.TenantID,
		CreatedAt:             comment.CreatedAt,
		CreatedByUserID:       comment.CreatedByUserID,
		CreatedByUserName:     comment.CreatedByUserName,
		CreatedFromIPAddress:  comment.CreatedFromIPAddress,
		ModifiedAt:            comment.ModifiedAt,
		ModifiedByUserID:      comment.ModifiedByUserID,
		ModifiedByUserName:    comment.ModifiedByUserName,
		ModifiedFromIPAddress: comment.ModifiedFromIPAddress,
		Text:                  comment.Text,
		Status:                comment.Status,
		OldID:                 comment.OldID,
	}

	// Append comments to associate details.
	associate.Comments = append(associate.Comments, cc)

	if err := custStorer.UpdateByID(ctx, associate); err != nil {
		log.Fatal(err)
	}

	//
	// Update the comment.
	//

	comment.BelongsTo = comm_ds.BelongsToAssociate
	comment.AssociateID = associate.ID
	comment.AssociateName = associate.Name
	if err := comStorer.UpdateByID(ctx, comment); err != nil {
		log.Fatal(err)
	}

	//
	// For debugging purposes only.
	//

	fmt.Println("Imported Associate Comment ID#", cc.ID, "for AssociateID", associate.ID)
}
