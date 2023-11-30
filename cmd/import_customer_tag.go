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
	cust_ds "github.com/over55/workery-cli/app/customer/datastore"
	hh_ds "github.com/over55/workery-cli/app/howhear/datastore"
	tag_ds "github.com/over55/workery-cli/app/tag/datastore"
	tenant_ds "github.com/over55/workery-cli/app/tenant/datastore"
	user_ds "github.com/over55/workery-cli/app/user/datastore"
	"github.com/over55/workery-cli/config"
)

func init() {
	rootCmd.AddCommand(importCustomerTagCmd)
}

var importCustomerTagCmd = &cobra.Command{
	Use:   "import_customer_tag",
	Short: "Import the customer tags from old database",
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
		tagStorer := tag_ds.NewDatastore(cfg, defaultLogger, mc)

		tenant, err := tenantStorer.GetBySchemaName(context.Background(), cfg.PostgresDB.DatabaseLondonSchemaName)
		if err != nil {
			log.Fatal(err)
		}

		RunImportCustomerTag(cfg, ppc, lpc, tenantStorer, userStorer, custStorer, hhStorer, tagStorer, tenant)
	},
}

type OldCustomerTag struct {
	Id         uint64 `json:"id"`
	CustomerId uint64 `json:"customer_id"`
	TagId      uint64 `json:"tag_id"`
}

func ListAllCustomerTags(db *sql.DB) ([]*OldCustomerTag, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `
	SELECT
	    id, customer_id, tag_id
	FROM
	    london.workery_customers_tags
	ORDER BY
	    id
	ASC
	`
	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}

	var arr []*OldCustomerTag
	defer rows.Close()
	for rows.Next() {
		m := new(OldCustomerTag)
		err = rows.Scan(
			&m.Id,
			&m.CustomerId,
			&m.TagId,
		)
		if err != nil {
			log.Panic("ListAllCustomerTags | Next | err", err)
		}
		arr = append(arr, m)
	}
	err = rows.Err()
	if err != nil {
		panic(err)
	}
	return arr, err
}

func RunImportCustomerTag(cfg *config.Conf, public *sql.DB, london *sql.DB, tenantStorer tenant_ds.TenantStorer, userStorer user_ds.UserStorer, custStorer cust_ds.CustomerStorer, hhStorer hh_ds.HowHearAboutUsItemStorer, tagStorer tag_ds.TagStorer, tenant *tenant_ds.Tenant) {
	fmt.Println("Beginning importing customers")
	data, err := ListAllCustomerTags(london)
	if err != nil {
		log.Fatal(err)
	}

	for _, datum := range data {
		importCustomerTag(context.Background(), tenantStorer, userStorer, custStorer, hhStorer, tagStorer, tenant, datum)
	}
	fmt.Println("Finished importing customers")
}

func importCustomerTag(ctx context.Context, ts tenant_ds.TenantStorer, us user_ds.UserStorer, custStorer cust_ds.CustomerStorer, hhStorer hh_ds.HowHearAboutUsItemStorer, tagStorer tag_ds.TagStorer, tenant *tenant_ds.Tenant, ou *OldCustomerTag) {

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
	tag, err := tagStorer.GetByOldID(ctx, ou.TagId)
	if err != nil {
		log.Fatal(err)
	}
	if tag == nil {
		log.Fatal("tag does not exist")
	}

	//
	// Create the customer tag.
	//

	cc := &cust_ds.CustomerTag{
		ID:          tag.ID,
		Text:        tag.Text,
		Description: tag.Description,
		Status:      tag.Status,
	}

	// Append tags to customer details.
	customer.Tags = append(customer.Tags, cc)

	if err := custStorer.UpdateByID(ctx, customer); err != nil {
		log.Fatal(err)
	}

	//
	// For debugging purposes only.
	//

	fmt.Println("Imported Customer Tag ID#", cc.ID, "for CustomerID", customer.ID)
}
