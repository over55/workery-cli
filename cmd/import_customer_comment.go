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
	cust_ds "github.com/over55/workery-cli/app/customer/datastore"
	hh_ds "github.com/over55/workery-cli/app/howhear/datastore"
	tenant_ds "github.com/over55/workery-cli/app/tenant/datastore"
	user_ds "github.com/over55/workery-cli/app/user/datastore"
	"github.com/over55/workery-cli/config"
)

func init() {
	rootCmd.AddCommand(importCustomerCommentCmd)
}

var importCustomerCommentCmd = &cobra.Command{
	Use:   "import_customer_comment",
	Short: "Import the customer comments from old database",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.New()
		mc := mongodb.NewStorage(cfg)
		ppc := postgres.NewStorage(cfg, cfg.PostgresDB.DatabasePublicSchemaName)
		lpc := postgres.NewStorage(cfg, cfg.PostgresDB.DatabasePublicSchemaName)
		defaultLogger := slog.Default()
		tenantStorer := tenant_ds.NewDatastore(cfg, defaultLogger, mc)
		userStorer := user_ds.NewDatastore(cfg, defaultLogger, mc)
		custStorer := cust_ds.NewDatastore(cfg, defaultLogger, mc)
		hhStorer := hh_ds.NewDatastore(cfg, defaultLogger, mc)
		comStorer := comm_ds.NewDatastore(cfg, defaultLogger, mc)

		tenant, err := tenantStorer.GetBySchemaName(context.Background(), cfg.PostgresDB.DatabaseLondonSchemaName)
		if err != nil {
			log.Fatal(err)
		}

		RunImportCustomerComment(cfg, ppc, lpc, tenantStorer, userStorer, custStorer, hhStorer, comStorer, tenant)
	},
}

type OldCustomerComment struct {
	Id         uint64    `json:"id"`
	CreatedAt  time.Time `json:"created_at"`
	CustomerId uint64    `json:"customer_id"`
	CommentId  uint64    `json:"comment_id"`
}

func ListAllCustomerComments(db *sql.DB) ([]*OldCustomerComment, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `
	SELECT
	    id, created_at, about_id, comment_id
	FROM
	    london.workery_customer_comments
	ORDER BY
	    id
	ASC
	`
	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}

	var arr []*OldCustomerComment
	defer rows.Close()
	for rows.Next() {
		m := new(OldCustomerComment)
		err = rows.Scan(
			&m.Id,
			&m.CreatedAt,
			&m.CustomerId,
			&m.CommentId,
		)
		if err != nil {
			log.Panic("ListAllCustomerComments | Next | err", err)
		}
		arr = append(arr, m)
	}
	err = rows.Err()
	if err != nil {
		panic(err)
	}
	return arr, err
}

func RunImportCustomerComment(cfg *config.Conf, public *sql.DB, london *sql.DB, tenantStorer tenant_ds.TenantStorer, userStorer user_ds.UserStorer, custStorer cust_ds.CustomerStorer, hhStorer hh_ds.HowHearAboutUsItemStorer, comStorer comm_ds.CommentStorer, tenant *tenant_ds.Tenant) {
	fmt.Println("Beginning importing customers")
	data, err := ListAllCustomerComments(london)
	if err != nil {
		log.Fatal(err)
	}

	for _, datum := range data {
		importCustomerComment(context.Background(), tenantStorer, userStorer, custStorer, hhStorer, comStorer, tenant, datum)
	}
	fmt.Println("Finished importing customers")
}

func importCustomerComment(ctx context.Context, ts tenant_ds.TenantStorer, us user_ds.UserStorer, custStorer cust_ds.CustomerStorer, hhStorer hh_ds.HowHearAboutUsItemStorer, comStorer comm_ds.CommentStorer, tenant *tenant_ds.Tenant, ou *OldCustomerComment) {

	//
	// Lookup related.
	//

	customer, err := custStorer.GetByOldID(ctx, ou.CustomerId)
	if err != nil {
		log.Fatal(err)
	}
	if customer == nil {
		log.Fatal("customer does not exist")
	}
	comment, err := comStorer.GetByOldID(ctx, ou.CommentId)
	if err != nil {
		log.Fatal(err)
	}
	if comment == nil {
		log.Fatal("comment does not exist")
	}

	//
	// Create the customer comment.
	//

	cc := &cust_ds.CustomerComment{
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

	// Append comments to customer details.
	customer.Comments = append(customer.Comments, cc)

	if err := custStorer.UpdateByID(ctx, customer); err != nil {
		log.Fatal(err)
	}

	//
	// Update the comment.
	//

	comment.BelongsTo = comm_ds.BelongsToCustomer
	comment.CustomerID = customer.ID
	comment.CustomerName = customer.Name
	if err := comStorer.UpdateByID(ctx, comment); err != nil {
		log.Fatal(err)
	}

	//
	// For debugging purposes only.
	//

	fmt.Println("Imported Customer Comment ID#", cc.ID, "for CustomerID", customer.ID)
}
