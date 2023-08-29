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
	comm_ds "github.com/over55/workery-cli/app/comment/datastore"
	hh_ds "github.com/over55/workery-cli/app/howhear/datastore"
	o_ds "github.com/over55/workery-cli/app/order/datastore"
	tenant_ds "github.com/over55/workery-cli/app/tenant/datastore"
	user_ds "github.com/over55/workery-cli/app/user/datastore"
	"github.com/over55/workery-cli/config"
)

func init() {
	rootCmd.AddCommand(importOrderCommentCmd)
}

var importOrderCommentCmd = &cobra.Command{
	Use:   "import_order_comment",
	Short: "Import the order comments from old database",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.New()
		mc := mongodb.NewStorage(cfg)
		ppc := postgres.NewStorage(cfg, cfg.PostgresDB.DatabasePublicSchemaName)
		lpc := postgres.NewStorage(cfg, cfg.PostgresDB.DatabasePublicSchemaName)
		defaultLogger := slog.Default()
		tenantStorer := tenant_ds.NewDatastore(cfg, defaultLogger, mc)
		userStorer := user_ds.NewDatastore(cfg, defaultLogger, mc)
		oStorer := o_ds.NewDatastore(cfg, defaultLogger, mc)
		hhStorer := hh_ds.NewDatastore(cfg, defaultLogger, mc)
		comStorer := comm_ds.NewDatastore(cfg, defaultLogger, mc)

		tenant, err := tenantStorer.GetBySchemaName(context.Background(), cfg.PostgresDB.DatabaseLondonSchemaName)
		if err != nil {
			log.Fatal(err)
		}

		RunImportOrderComment(cfg, ppc, lpc, tenantStorer, userStorer, oStorer, hhStorer, comStorer, tenant)
	},
}

type OldWorkOrderComment struct {
	Id          uint64    `json:"id"`
	CreatedAt   time.Time `json:"created_at"`
	WorkOrderId uint64    `json:"about_id"`
	CommentId   uint64    `json:"comment_id"`
}

func ListAllWorkOrderComments(db *sql.DB) ([]*OldWorkOrderComment, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `
	SELECT
        id, created_at, about_id, comment_id
	FROM
        london.workery_work_order_comments
	ORDER BY
	    id
	ASC
	`
	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}

	var arr []*OldWorkOrderComment
	defer rows.Close()
	for rows.Next() {
		m := new(OldWorkOrderComment)
		err = rows.Scan(
			&m.Id,
			&m.CreatedAt,
			&m.WorkOrderId,
			&m.CommentId,
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

func RunImportOrderComment(cfg *config.Conf, public *sql.DB, london *sql.DB, tenantStorer tenant_ds.TenantStorer, userStorer user_ds.UserStorer, oStorer o_ds.OrderStorer, hhStorer hh_ds.HowHearAboutUsItemStorer, comStorer comm_ds.CommentStorer, tenant *tenant_ds.Tenant) {
	fmt.Println("Beginning importing associates")
	data, err := ListAllWorkOrderComments(london)
	if err != nil {
		log.Fatal(err)
	}

	for _, datum := range data {
		importOrderComment(context.Background(), tenantStorer, userStorer, oStorer, hhStorer, comStorer, tenant, datum)
	}
	fmt.Println("Finished importing associates")
}

func importOrderComment(ctx context.Context, ts tenant_ds.TenantStorer, us user_ds.UserStorer, oStorer o_ds.OrderStorer, hhStorer hh_ds.HowHearAboutUsItemStorer, comStorer comm_ds.CommentStorer, tenant *tenant_ds.Tenant, ou *OldWorkOrderComment) {

	//
	// Lookup related.
	//

	order, err := oStorer.GetByOldID(ctx, ou.WorkOrderId)
	if err != nil {
		log.Fatal(err)
	}
	if order == nil {
		log.Fatal("order does not exist")
	}
	comment, err := comStorer.GetByOldID(ctx, ou.CommentId)
	if err != nil {
		log.Fatal(err)
	}
	if comment == nil {
		log.Fatal("comment does not exist")
	}

	//
	// Create the order comment.
	//

	oc := &o_ds.OrderComment{
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
		Content:               comment.Content,
		Status:                comment.Status,
		OldID:                 comment.OldID,
	}

	// Append comments to order details.
	order.Comments = append(order.Comments, oc)

	if err := oStorer.UpdateByID(ctx, order); err != nil {
		log.Fatal(err)
	}

	//
	// Update the comment.
	//

	comment.BelongsTo = comm_ds.BelongsToOrder
	comment.OrderID = order.ID
	if err := comStorer.UpdateByID(ctx, comment); err != nil {
		log.Fatal(err)
	}

	//
	// For debugging purposes only.
	//

	fmt.Println("Imported Order Comment ID#", oc.ID, "for OrderID", order.ID)
}
