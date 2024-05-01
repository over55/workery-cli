package cmd

import (
	"context"
	"fmt"
	"log"
	"log/slog"

	"github.com/over55/workery-cli/adapter/storage/mongodb"
	a_ds "github.com/over55/workery-cli/app/associate/datastore"
	c_ds "github.com/over55/workery-cli/app/customer/datastore"
	o_ds "github.com/over55/workery-cli/app/order/datastore"
	s_ds "github.com/over55/workery-cli/app/staff/datastore"
	ti_ds "github.com/over55/workery-cli/app/taskitem/datastore"
	tenant_ds "github.com/over55/workery-cli/app/tenant/datastore"
	user_ds "github.com/over55/workery-cli/app/user/datastore"
	"github.com/over55/workery-cli/config"
	"github.com/spf13/cobra"
	"go.mongodb.org/mongo-driver/mongo"
)

func init() {
	rootCmd.AddCommand(hotfix05Cmd)
}

var hotfix05Cmd = &cobra.Command{
	Use:   "hotfix05",
	Short: "Execute hotfix05",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("hotfix05")

		cfg := config.New()
		mc := mongodb.NewStorage(cfg)
		defaultLogger := slog.Default()
		tenantStorer := tenant_ds.NewDatastore(cfg, defaultLogger, mc)
		userStorer := user_ds.NewDatastore(cfg, defaultLogger, mc)
		cStorer := c_ds.NewDatastore(cfg, defaultLogger, mc)
		aStorer := a_ds.NewDatastore(cfg, defaultLogger, mc)
		sStorer := s_ds.NewDatastore(cfg, defaultLogger, mc)
		oStorer := o_ds.NewDatastore(cfg, defaultLogger, mc)
		tiStorer := ti_ds.NewDatastore(cfg, defaultLogger, mc)
		tenant, err := tenantStorer.GetBySchemaName(context.Background(), cfg.PostgresDB.DatabaseLondonSchemaName)
		if err != nil {
			log.Fatal(err)
		}

		RunHotfix05(cfg, mc, tenantStorer, userStorer, cStorer, aStorer, sStorer, oStorer, tiStorer, tenant)

	},
}

func RunHotfix05(
	cfg *config.Conf,
	mc *mongo.Client,
	tenantStorer tenant_ds.TenantStorer,
	userStorer user_ds.UserStorer,
	cStorer c_ds.CustomerStorer,
	aStorer a_ds.AssociateStorer,
	sStorer s_ds.StaffStorer,
	oStorer o_ds.OrderStorer,
	tiStorer ti_ds.TaskItemStorer,
	tenant *tenant_ds.Tenant,
) {
	session, err := mc.StartSession()
	if err != nil {
		log.Fatal(err)
	}
	defer session.EndSession(context.Background())

	// Define a transaction function with a series of operations
	transactionFunc := func(sessCtx mongo.SessionContext) (interface{}, error) {
		aa, err := aStorer.ListAll(context.Background())
		if err != nil {
			return nil, err
		}
		for _, a := range aa.Results {
			oo, err := oStorer.ListByAssociateID(context.Background(), a.ID)
			if err != nil {
				return nil, err
			}
			for _, o := range oo.Results {
				log.Println("A_ID", a.ID, "| O_ID", o.ID)
				o.AssociateServiceFeeID = a.ServiceFeeID
				o.AssociateServiceFeeName = a.ServiceFeeName
				o.AssociateServiceFeePercentage = a.ServiceFeePercentage
				if err := oStorer.UpdateByID(context.Background(), o); err != nil {
					return nil, err
				}
			}

			titi, err := tiStorer.ListByAssociateID(context.Background(), a.ID)
			if err != nil {
				return nil, err
			}
			for _, ti := range titi.Results {
				log.Println("A_ID", a.ID, "| TI_ID", ti.ID)
				ti.AssociateServiceFeeID = a.ServiceFeeID
				ti.AssociateServiceFeeName = a.ServiceFeeName
				ti.AssociateServiceFeePercentage = a.ServiceFeePercentage
				if err := tiStorer.UpdateByID(context.Background(), ti); err != nil {
					return nil, err
				}
			}

		}
		return nil, nil
	}

	// Start a transaction
	if _, err := session.WithTransaction(context.Background(), transactionFunc); err != nil {
		log.Fatal(err)
	}
}
