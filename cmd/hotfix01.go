package cmd

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"log/slog"

	"github.com/over55/workery-cli/adapter/storage/mongodb"
	"github.com/over55/workery-cli/adapter/storage/postgres"
	a_ds "github.com/over55/workery-cli/app/associate/datastore"
	c_ds "github.com/over55/workery-cli/app/customer/datastore"
	hh_ds "github.com/over55/workery-cli/app/howhear/datastore"
	s_ds "github.com/over55/workery-cli/app/staff/datastore"
	tenant_ds "github.com/over55/workery-cli/app/tenant/datastore"
	user_ds "github.com/over55/workery-cli/app/user/datastore"
	"github.com/over55/workery-cli/config"
	"github.com/spf13/cobra"
	"go.mongodb.org/mongo-driver/mongo"
)

func init() {
	rootCmd.AddCommand(hotfix1Cmd)
}

var hotfix1Cmd = &cobra.Command{
	Use:   "hotfix01",
	Short: "Execute hotfix01",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("hotfix01")

		cfg := config.New()
		mc := mongodb.NewStorage(cfg)
		ppc := postgres.NewStorage(cfg, cfg.PostgresDB.DatabasePublicSchemaName)
		lpc := postgres.NewStorage(cfg, cfg.PostgresDB.DatabasePublicSchemaName)
		defaultLogger := slog.Default()
		tenantStorer := tenant_ds.NewDatastore(cfg, defaultLogger, mc)
		userStorer := user_ds.NewDatastore(cfg, defaultLogger, mc)
		cStorer := c_ds.NewDatastore(cfg, defaultLogger, mc)
		aStorer := a_ds.NewDatastore(cfg, defaultLogger, mc)
		sStorer := s_ds.NewDatastore(cfg, defaultLogger, mc)
		hhStorer := hh_ds.NewDatastore(cfg, defaultLogger, mc)

		tenant, err := tenantStorer.GetBySchemaName(context.Background(), cfg.PostgresDB.DatabaseLondonSchemaName)
		if err != nil {
			log.Fatal(err)
		}

		RunHotfix01(cfg, ppc, lpc, mc, tenantStorer, userStorer, cStorer, aStorer, sStorer, hhStorer, tenant)

	},
}

func RunHotfix01(
	cfg *config.Conf,
	public *sql.DB,
	london *sql.DB,
	mc *mongo.Client,
	tenantStorer tenant_ds.TenantStorer,
	userStorer user_ds.UserStorer,
	cStorer c_ds.CustomerStorer,
	aStorer a_ds.AssociateStorer,
	sStorer s_ds.StaffStorer,
	hhStorer hh_ds.HowHearAboutUsItemStorer,
	tenant *tenant_ds.Tenant,
) {
	fmt.Println("Beginning importing customers")
	customerData, err := ListAllCustomers(london)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Beginning importing associates")
	associateData, err := ListAllAssociates(london)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Beginning importing staffs")
	staffData, err := ListAllStaffs(london)
	if err != nil {
		log.Fatal(err)
	}

	for _, datum := range customerData {
		if err := hotfix01Customer(mc, tenantStorer, userStorer, cStorer, hhStorer, tenant, datum); err != nil {
			log.Fatal(err)
		}
	}
	fmt.Println("Finished importing associates")
	for _, datum := range associateData {
		if err := hotfix01Associate(mc, tenantStorer, userStorer, aStorer, hhStorer, tenant, datum); err != nil {
			log.Fatal(err)
		}
	}
	fmt.Println("Finished importing associates")
	for _, datum := range staffData {
		if err := hotfix01Staff(mc, tenantStorer, userStorer, sStorer, hhStorer, tenant, datum); err != nil {
			log.Fatal(err)
		}
	}
	fmt.Println("Finished importing staffs")
}

func hotfix01Customer(
	mc *mongo.Client,
	ts tenant_ds.TenantStorer,
	userStorer user_ds.UserStorer,
	customerStorer c_ds.CustomerStorer,
	hhStorer hh_ds.HowHearAboutUsItemStorer,
	tenant *tenant_ds.Tenant,
	oldCustomer *OldCustomer,
) error {
	////
	//// Start the transaction.
	////

	session, err := mc.StartSession()
	if err != nil {
		log.Fatal(err)
	}
	defer session.EndSession(context.Background())

	// Define a transaction function with a series of operations
	transactionFunc := func(sessCtx mongo.SessionContext) (interface{}, error) {
		newCustomer, err := customerStorer.GetByEmail(sessCtx, oldCustomer.Email.ValueOrZero())
		if err != nil {
			return nil, err
		}
		if newCustomer == nil {
			err := fmt.Errorf("does not exist for Email: %v", oldCustomer.Email.ValueOrZero())
			log.Println("err:", err)
			return nil, nil
			// return err
		}

		//
		// Fix 1 - Make sure correct public id is set.
		//

		newCustomer.PublicID = oldCustomer.ID

		// // for debugging purposes.
		// fmt.Println(newCustomer.PublicID, oldCustomer.ID, "--->", newCustomer.Email)

		// Save fix 1.
		if err := customerStorer.UpdateByID(sessCtx, newCustomer); err != nil {
			return nil, err
		}

		//
		// Fix 2 - Modified by
		//

		modifiedByCustomer, err := userStorer.GetByPublicID(sessCtx, uint64(oldCustomer.LastModifiedById.ValueOrZero()))
		if err != nil {
			return nil, err
		}
		if modifiedByCustomer == nil {
			err := fmt.Errorf("does not exist for LastModifiedById: %v", oldCustomer.LastModifiedById.ValueOrZero())
			log.Println("err:", err)
			return nil, nil
			// return err
		}

		newCustomer.ModifiedByUserID = modifiedByCustomer.ID
		newCustomer.ModifiedByUserName = modifiedByCustomer.Name

		// fmt.Println("newCustomer.ModifiedByUserID -->", newCustomer.ID)
		// fmt.Println(newCustomer.ID, newCustomer.Name, "-->", newCustomer.ModifiedByUserName)
		// fmt.Println()

		// Save fix 2.
		if err := customerStorer.UpdateByID(sessCtx, newCustomer); err != nil {
			return nil, err
		}

		fmt.Println("hotfixed customer:", newCustomer.PublicID)
		return nil, nil
	}

	// Start a transaction
	if _, err := session.WithTransaction(context.Background(), transactionFunc); err != nil {
		log.Fatal(err)
	}
	return nil
}

func hotfix01Associate(
	mc *mongo.Client,
	ts tenant_ds.TenantStorer,
	userStorer user_ds.UserStorer,
	associateStorer a_ds.AssociateStorer,
	hhStorer hh_ds.HowHearAboutUsItemStorer,
	tenant *tenant_ds.Tenant,
	oldAssociate *OldAssociate,
) error {
	////
	//// Start the transaction.
	////

	session, err := mc.StartSession()
	if err != nil {
		log.Fatal(err)
	}
	defer session.EndSession(context.Background())

	// Define a transaction function with a series of operations
	transactionFunc := func(sessCtx mongo.SessionContext) (interface{}, error) {
		newAssociate, err := associateStorer.GetByEmail(sessCtx, oldAssociate.Email.ValueOrZero())
		if err != nil {
			return nil, err
		}
		if newAssociate == nil {
			err := fmt.Errorf("does not exist for Email: %v", oldAssociate.Email.ValueOrZero())
			log.Println("err:", err)
			return nil, nil // This is not an mistake, we simply want to stop this function.
		}

		//
		// Fix 1 - Make sure correct public id is set.
		//

		newAssociate.PublicID = oldAssociate.ID

		// // for debugging purposes.
		// fmt.Println(newAssociate.PublicID, oldAssociate.ID, "--->", newAssociate.Email)

		// Save fix 1.
		if err := associateStorer.UpdateByID(sessCtx, newAssociate); err != nil {
			return nil, err
		}

		//
		// Fix 2 - Modified by
		//

		modifiedByAssociate, err := userStorer.GetByPublicID(sessCtx, uint64(oldAssociate.LastModifiedByID.ValueOrZero()))
		if err != nil {
			return nil, err
		}
		if modifiedByAssociate == nil {
			err := fmt.Errorf("does not exist for LastModifiedByID: %v", oldAssociate.LastModifiedByID.ValueOrZero())
			log.Println("err:", err)
			return nil, nil // This is not an mistake, we simply want to stop this function.
		}

		newAssociate.ModifiedByUserID = modifiedByAssociate.ID
		newAssociate.ModifiedByUserName = modifiedByAssociate.Name

		// fmt.Println("newAssociate.ModifiedByUserID -->", newAssociate.ID)
		// fmt.Println(newAssociate.ID, newAssociate.Name, "-->", newAssociate.ModifiedByUserName)
		// fmt.Println()

		// Save fix 2.
		if err := associateStorer.UpdateByID(sessCtx, newAssociate); err != nil {
			return nil, err
		}

		fmt.Println("hotfixed associate:", newAssociate.PublicID)
		return nil, nil
	}

	// Start a transaction
	if _, err := session.WithTransaction(context.Background(), transactionFunc); err != nil {
		log.Fatal(err)
	}
	return nil
}

func hotfix01Staff(
	mc *mongo.Client,
	ts tenant_ds.TenantStorer,
	userStorer user_ds.UserStorer,
	staffStorer s_ds.StaffStorer,
	hhStorer hh_ds.HowHearAboutUsItemStorer,
	tenant *tenant_ds.Tenant,
	oldStaff *OldStaff,
) error {
	////
	//// Start the transaction.
	////

	session, err := mc.StartSession()
	if err != nil {
		log.Fatal(err)
	}
	defer session.EndSession(context.Background())

	// Define a transaction function with a series of operations
	transactionFunc := func(sessCtx mongo.SessionContext) (interface{}, error) {
		newStaff, err := staffStorer.GetByEmail(sessCtx, oldStaff.Email.ValueOrZero())
		if err != nil {
			return nil, err
		}
		if newStaff == nil {
			err := fmt.Errorf("does not exist for Email: %v", oldStaff.Email.ValueOrZero())
			log.Println("err:", err)
			return nil, nil // This is not an mistake, we simply want to stop this function.
		}

		//
		// Fix 1 - Make sure correct public id is set.
		//

		newStaff.PublicID = oldStaff.ID

		// // for debugging purposes.
		// fmt.Println(newStaff.PublicID, oldStaff.ID, "--->", newStaff.Email)

		// Save fix 1.
		if err := staffStorer.UpdateByID(sessCtx, newStaff); err != nil {
			return nil, err
		}

		//
		// Fix 2 - Modified by
		//

		modifiedByStaff, err := userStorer.GetByPublicID(sessCtx, uint64(oldStaff.LastModifiedByID.ValueOrZero()))
		if err != nil {
			return nil, err
		}
		if modifiedByStaff == nil {
			err := fmt.Errorf("does not exist for LastModifiedByID: %v", oldStaff.LastModifiedByID.ValueOrZero())
			log.Println("err:", err)
			return nil, nil // This is not an mistake, we simply want to stop this function.
		}

		newStaff.ModifiedByUserID = modifiedByStaff.ID
		newStaff.ModifiedByUserName = modifiedByStaff.Name

		// fmt.Println("newStaff.ModifiedByUserID -->", newStaff.ID)
		// fmt.Println(newStaff.ID, newStaff.Name, "-->", newStaff.ModifiedByUserName)
		// fmt.Println()

		// Save fix 2.
		if err := staffStorer.UpdateByID(sessCtx, newStaff); err != nil {
			return nil, err
		}

		fmt.Println("hotfixed staff:", newStaff.PublicID)
		return nil, nil
	}

	// Start a transaction
	if _, err := session.WithTransaction(context.Background(), transactionFunc); err != nil {
		log.Fatal(err)
	}
	return nil
}
