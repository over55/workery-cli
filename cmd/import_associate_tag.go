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
	cust_ds "github.com/over55/workery-cli/app/associate/datastore"
	hh_ds "github.com/over55/workery-cli/app/howhear/datastore"
	tag_ds "github.com/over55/workery-cli/app/tag/datastore"
	tenant_ds "github.com/over55/workery-cli/app/tenant/datastore"
	user_ds "github.com/over55/workery-cli/app/user/datastore"
	"github.com/over55/workery-cli/config"
)

func init() {
	rootCmd.AddCommand(importAssociateTagCmd)
}

var importAssociateTagCmd = &cobra.Command{
	Use:   "import_associate_tag",
	Short: "Import the associate tags from old database",
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

		RunImportAssociateTag(cfg, ppc, lpc, tenantStorer, userStorer, custStorer, hhStorer, tagStorer, tenant)
	},
}

type OldAssociateTag struct {
	Id          uint64 `json:"id"`
	AssociateId uint64 `json:"associate_id"`
	TagId       uint64 `json:"tag_id"`
}

func ListAllAssociateTags(db *sql.DB) ([]*OldAssociateTag, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `
	SELECT
	    id, associate_id, tag_id
	FROM
	    london.workery_associates_tags
	ORDER BY
	    id
	ASC
	`
	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}

	var arr []*OldAssociateTag
	defer rows.Close()
	for rows.Next() {
		m := new(OldAssociateTag)
		err = rows.Scan(
			&m.Id,
			&m.AssociateId,
			&m.TagId,
		)
		if err != nil {
			log.Panic("ListAllAssociateTags | Next | err", err)
		}
		arr = append(arr, m)
	}
	err = rows.Err()
	if err != nil {
		panic(err)
	}
	return arr, err
}

func RunImportAssociateTag(cfg *config.Conf, public *sql.DB, london *sql.DB, tenantStorer tenant_ds.TenantStorer, userStorer user_ds.UserStorer, custStorer cust_ds.AssociateStorer, hhStorer hh_ds.HowHearAboutUsItemStorer, tagStorer tag_ds.TagStorer, tenant *tenant_ds.Tenant) {
	fmt.Println("Beginning importing associates")
	data, err := ListAllAssociateTags(london)
	if err != nil {
		log.Fatal(err)
	}

	for _, datum := range data {
		importAssociateTag(context.Background(), tenantStorer, userStorer, custStorer, hhStorer, tagStorer, tenant, datum)
	}
	fmt.Println("Finished importing associates")
}

func importAssociateTag(ctx context.Context, ts tenant_ds.TenantStorer, us user_ds.UserStorer, custStorer cust_ds.AssociateStorer, hhStorer hh_ds.HowHearAboutUsItemStorer, tagStorer tag_ds.TagStorer, tenant *tenant_ds.Tenant, ou *OldAssociateTag) {

	//
	// Lookup related.
	//

	associate, err := custStorer.GetByPublicID(ctx, ou.AssociateId)
	if err != nil {
		log.Fatal(err)
	}
	if associate == nil {
		log.Fatal("associate does not exist")
	}
	tag, err := tagStorer.GetByPublicID(ctx, ou.TagId)
	if err != nil {
		log.Fatal(err)
	}
	if tag == nil {
		log.Fatal("tag does not exist")
	}

	//
	// Create the associate tag.
	//

	cc := &cust_ds.AssociateTag{
		ID:          tag.ID,
		Text:        tag.Text,
		Description: tag.Description,
		Status:      tag.Status,
	}

	// Append tags to associate details.
	associate.Tags = append(associate.Tags, cc)

	if err := custStorer.UpdateByID(ctx, associate); err != nil {
		log.Fatal(err)
	}

	//
	// For debugging purposes only.
	//

	fmt.Println("Imported Associate Tag ID#", cc.ID, "for AssociateID", associate.ID)
}
