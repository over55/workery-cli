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
	hh_ds "github.com/over55/workery-cli/app/howhear/datastore"
	o_ds "github.com/over55/workery-cli/app/order/datastore"
	comm_ds "github.com/over55/workery-cli/app/tag/datastore"
	tenant_ds "github.com/over55/workery-cli/app/tenant/datastore"
	user_ds "github.com/over55/workery-cli/app/user/datastore"
	"github.com/over55/workery-cli/config"
)

func init() {
	rootCmd.AddCommand(importOrderTagCmd)
}

var importOrderTagCmd = &cobra.Command{
	Use:   "import_order_tag",
	Short: "Import the order tags from old database",
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

		RunImportOrderTag(cfg, ppc, lpc, tenantStorer, userStorer, oStorer, hhStorer, comStorer, tenant)
	},
}

type OldWorkOrderTag struct {
	Id          uint64    `json:"id"`
	CreatedAt   time.Time `json:"created_at"`
	WorkOrderId uint64    `json:"about_id"`
	TagId       uint64    `json:"tag_id"`
}

func ListAllWorkOrderTags(db *sql.DB) ([]*OldWorkOrderTag, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `
	SELECT
        id, workorder_id, tag_id
	FROM
        london.workery_work_orders_tags
	ORDER BY
	    id
	ASC
	`
	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}

	var arr []*OldWorkOrderTag
	defer rows.Close()
	for rows.Next() {
		m := new(OldWorkOrderTag)
		err = rows.Scan(
			&m.Id,
			&m.WorkOrderId,
			&m.TagId,
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

func RunImportOrderTag(cfg *config.Conf, public *sql.DB, london *sql.DB, tenantStorer tenant_ds.TenantStorer, userStorer user_ds.UserStorer, oStorer o_ds.OrderStorer, hhStorer hh_ds.HowHearAboutUsItemStorer, comStorer comm_ds.TagStorer, tenant *tenant_ds.Tenant) {
	fmt.Println("Beginning importing order tags")
	data, err := ListAllWorkOrderTags(london)
	if err != nil {
		log.Fatal(err)
	}

	for _, datum := range data {
		importOrderTag(context.Background(), tenantStorer, userStorer, oStorer, hhStorer, comStorer, tenant, datum)
	}
	fmt.Println("Finished importing order tags")
}

func importOrderTag(ctx context.Context, ts tenant_ds.TenantStorer, us user_ds.UserStorer, oStorer o_ds.OrderStorer, hhStorer hh_ds.HowHearAboutUsItemStorer, comStorer comm_ds.TagStorer, tenant *tenant_ds.Tenant, ou *OldWorkOrderTag) {

	//
	// Lookup related.
	//

	order, err := oStorer.GetByWJID(ctx, ou.WorkOrderId)
	if err != nil {
		log.Fatal(err)
	}
	if order == nil {
		log.Fatal("order does not exist")
	}
	tag, err := comStorer.GetByOldID(ctx, ou.TagId)
	if err != nil {
		log.Fatal(err)
	}
	if tag == nil {
		log.Fatal("tag does not exist")
	}

	//
	// Create the order tag.
	//

	oc := &o_ds.OrderTag{
		ID:          tag.ID,
		OrderID:     order.ID,
		OrderWJID:   order.WJID,
		TenantID:    tag.TenantID,
		Text:        tag.Text,
		Description: tag.Description,
		OldID:       tag.OldID,
	}

	// Append tags to order details.
	order.Tags = append(order.Tags, oc)

	if err := oStorer.UpdateByID(ctx, order); err != nil {
		log.Fatal(err)
	}

	//
	// For debugging purposes only.
	//

	fmt.Println("Imported Order Tag ID#", oc.ID, "for OrderID", order.ID)
}
