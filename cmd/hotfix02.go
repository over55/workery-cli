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
	attachment "github.com/over55/workery-cli/app/attachment/datastore"
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
	rootCmd.AddCommand(hotfix02Cmd)
}

var hotfix02Cmd = &cobra.Command{
	Use:   "hotfix02",
	Short: "Execute hotfix02",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("hotfix02")

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
		attachStorer := attachment.NewDatastore(cfg, defaultLogger, mc)

		tenant, err := tenantStorer.GetBySchemaName(context.Background(), cfg.PostgresDB.DatabaseLondonSchemaName)
		if err != nil {
			log.Fatal(err)
		}

		RunHotfix02(cfg, ppc, lpc, mc, tenantStorer, userStorer, cStorer, aStorer, sStorer, hhStorer, attachStorer, tenant)

	},
}

func RunHotfix02(
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
	attachStorer attachment.AttachmentStorer,
	tenant *tenant_ds.Tenant,
) {
	if err := hotfix02Customer(mc, tenantStorer, userStorer, cStorer, hhStorer, attachStorer, tenant); err != nil {
		log.Fatal(err)
	}
	if err := hotfix02Associate(mc, tenantStorer, userStorer, aStorer, hhStorer, attachStorer, tenant); err != nil {
		log.Fatal(err)
	}
	if err := hotfix02Staff(mc, tenantStorer, userStorer, sStorer, hhStorer, attachStorer, tenant); err != nil {
		log.Fatal(err)
	}
}

func hotfix02Customer(
	mc *mongo.Client,
	ts tenant_ds.TenantStorer,
	userStorer user_ds.UserStorer,
	customerStorer c_ds.CustomerStorer,
	hhStorer hh_ds.HowHearAboutUsItemStorer,
	attachStorer attachment.AttachmentStorer,
	tenant *tenant_ds.Tenant,
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

		// STEP 1: Get all the attachments in our system that belong to
		// customers only.
		res, err := attachStorer.ListByType(sessCtx, attachment.AttachmentTypeCustomer)
		if err != nil {
			return nil, err
		}

		// STEP 2: Group all the attachments per customer.
		customerAttachments := make(map[string][]*attachment.Attachment)
		for _, a := range res.Results {
			// // For debugging purposes only.
			// log.Println("ID:", a.ID)
			// log.Println("CustomerID:", a.CustomerID)
			// log.Println("Filename:", a.Filename)
			// log.Println("ObjectKey:", a.ObjectKey)
			// log.Println("ObjectURL:", a.ObjectURL)
			// log.Println()

			// Extract the array of attachments that belong to a particular
			// customer user.
			arr := customerAttachments[a.CustomerID.Hex()]

			// Add attachment to the customer user's attachments.
			arr = append(arr, a)

			// Update our customer user's attachments array.
			customerAttachments[a.CustomerID.Hex()] = arr
		}

		// log.Println(customerAttachments) // For debugging purposes only.

		// STEP 3: Iterate through all the customers.
		for _, attachments := range customerAttachments {
			// log.Println(customerID, attachments) // For debugging purposes only.

			// STEP 4: Iterate over all the attachments and group similar files.
			similarAttachments := make(map[string][]*attachment.Attachment)
			for _, a := range attachments {

				// Extract the array of attachments that are similar
				arr := similarAttachments[a.ObjectKey]

				// Add attachment to the similar attachments.
				arr = append(arr, a)

				// Update our customer user's attachments array.
				similarAttachments[a.ObjectKey] = arr
			}
			// log.Println(similarAttachments) // For debugging purposes only.

			// DEVELOPERS NOTE:
			// Once we are here then we need to (1) pick one attachment and
			// (2) delete remaining similar attachments.

			// STEP 5: Iterate through all similar files.
			for objectKey, attachmentsByObjectKey := range similarAttachments {
				log.Println("////////////////////////////////////////////////////////////////////////////////////////////////////////////")
				log.Println("////////////////////////////////////////////////////////////////////////////////////////////////////////////")
				log.Println("////////////////////////////////////////////////////////////////////////////////////////////////////////////")
				log.Println("objectKey:", objectKey)
				log.Println("------------------------------------------------------------------------------------------------------------")

				var uniqueAttachment *attachment.Attachment
				var duplicateAttachments []*attachment.Attachment

				for _, a := range attachmentsByObjectKey {
					if uniqueAttachment == nil {
						uniqueAttachment = a
					} else {
						duplicateAttachments = append(duplicateAttachments, a)
					}

				}

				// STEP 6: Iterate over duplicates and delete RECORD ONLY, DO
				// NOT DELETE FILE IN S3!

				log.Println("keeping --->", uniqueAttachment.ObjectKey, uniqueAttachment.ID)
				for _, dup := range duplicateAttachments {

					// For defensive code purposes, just do a few tests to
					// make sure the record is similar before proceeding to
					// delete.
					if uniqueAttachment.CustomerID == dup.CustomerID && uniqueAttachment.ObjectKey == dup.ObjectKey {
						log.Println("remove --->", dup.ObjectKey, dup.ID)
						if err := attachStorer.DeleteByID(sessCtx, dup.ID); err != nil {
							log.Println("error deleting customer attachment:", err)
							return nil, err
						}
					}
				}
				//------ END
			}
		}

		return nil, nil
	}

	// Start a transaction
	if _, err := session.WithTransaction(context.Background(), transactionFunc); err != nil {
		log.Fatal(err)
	}
	return nil
}

func hotfix02Associate(
	mc *mongo.Client,
	ts tenant_ds.TenantStorer,
	userStorer user_ds.UserStorer,
	associateStorer a_ds.AssociateStorer,
	hhStorer hh_ds.HowHearAboutUsItemStorer,
	attachStorer attachment.AttachmentStorer,
	tenant *tenant_ds.Tenant,
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
		// STEP 1: Get all the attachments in our system that belong to
		// associates only.
		res, err := attachStorer.ListByType(sessCtx, attachment.AttachmentTypeAssociate)
		if err != nil {
			return nil, err
		}

		// STEP 2: Group all the attachments per associate.
		associateAttachments := make(map[string][]*attachment.Attachment)
		for _, a := range res.Results {
			// // For debugging purposes only.
			// log.Println("ID:", a.ID)
			// log.Println("AssociateID:", a.AssociateID)
			// log.Println("Filename:", a.Filename)
			// log.Println("ObjectKey:", a.ObjectKey)
			// log.Println("ObjectURL:", a.ObjectURL)
			// log.Println()

			// Extract the array of attachments that belong to a particular
			// associate user.
			arr := associateAttachments[a.AssociateID.Hex()]

			// Add attachment to the associate user's attachments.
			arr = append(arr, a)

			// Update our associate user's attachments array.
			associateAttachments[a.AssociateID.Hex()] = arr
		}

		// log.Println(associateAttachments) // For debugging purposes only.

		// STEP 3: Iterate through all the associates.
		for _, attachments := range associateAttachments {
			// log.Println(associateID, attachments) // For debugging purposes only.

			// STEP 4: Iterate over all the attachments and group similar files.
			similarAttachments := make(map[string][]*attachment.Attachment)
			for _, a := range attachments {

				// Extract the array of attachments that are similar
				arr := similarAttachments[a.ObjectKey]

				// Add attachment to the similar attachments.
				arr = append(arr, a)

				// Update our associate user's attachments array.
				similarAttachments[a.ObjectKey] = arr
			}
			log.Println(similarAttachments) // For debugging purposes only.

			// DEVELOPERS NOTE:
			// Once we are here then we need to (1) pick one attachment and
			// (2) delete remaining similar attachments.

			// STEP 5: Iterate through all similar files.
			for objectKey, attachmentsByObjectKey := range similarAttachments {
				log.Println("////////////////////////////////////////////////////////////////////////////////////////////////////////////")
				log.Println("////////////////////////////////////////////////////////////////////////////////////////////////////////////")
				log.Println("////////////////////////////////////////////////////////////////////////////////////////////////////////////")
				log.Println("objectKey:", objectKey)
				log.Println("------------------------------------------------------------------------------------------------------------")

				var uniqueAttachment *attachment.Attachment
				var duplicateAttachments []*attachment.Attachment

				for _, a := range attachmentsByObjectKey {
					if uniqueAttachment == nil {
						uniqueAttachment = a
					} else {
						duplicateAttachments = append(duplicateAttachments, a)
					}

				}

				// STEP 6: Iterate over duplicates and delete RECORD ONLY, DO
				// NOT DELETE FILE IN S3!

				log.Println("keeping --->", uniqueAttachment.ObjectKey, uniqueAttachment.ID)
				for _, dup := range duplicateAttachments {

					// For defensive code purposes, just do a few tests to
					// make sure the record is similar before proceeding to
					// delete.
					if uniqueAttachment.AssociateID == dup.AssociateID && uniqueAttachment.ObjectKey == dup.ObjectKey {
						log.Println("remove --->", dup.ObjectKey, dup.ID)
						if err := attachStorer.DeleteByID(sessCtx, dup.ID); err != nil {
							log.Println("error deleting associate attachment:", err)
							return nil, err
						}
					}
				}
			}
			//------ END
		}

		return nil, nil
	}

	// Start a transaction
	if _, err := session.WithTransaction(context.Background(), transactionFunc); err != nil {
		log.Fatal(err)
	}
	return nil
}

func hotfix02Staff(
	mc *mongo.Client,
	ts tenant_ds.TenantStorer,
	userStorer user_ds.UserStorer,
	staffStorer s_ds.StaffStorer,
	hhStorer hh_ds.HowHearAboutUsItemStorer,
	attachStorer attachment.AttachmentStorer,
	tenant *tenant_ds.Tenant,
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
		// STEP 1: Get all the attachments in our system that belong to
		// staffs only.
		res, err := attachStorer.ListByType(sessCtx, attachment.AttachmentTypeStaff)
		if err != nil {
			return nil, err
		}

		// STEP 2: Group all the attachments per staff.
		staffAttachments := make(map[string][]*attachment.Attachment)
		for _, a := range res.Results {
			// // For debugging purposes only.
			// log.Println("ID:", a.ID)
			// log.Println("StaffID:", a.StaffID)
			// log.Println("Filename:", a.Filename)
			// log.Println("ObjectKey:", a.ObjectKey)
			// log.Println("ObjectURL:", a.ObjectURL)
			// log.Println()

			// Extract the array of attachments that belong to a particular
			// staff user.
			arr := staffAttachments[a.StaffID.Hex()]

			// Add attachment to the staff user's attachments.
			arr = append(arr, a)

			// Update our staff user's attachments array.
			staffAttachments[a.StaffID.Hex()] = arr
		}

		// log.Println(staffAttachments) // For debugging purposes only.

		// STEP 3: Iterate through all the staffs.
		for _, attachments := range staffAttachments {
			// log.Println(staffID, attachments) // For debugging purposes only.

			// STEP 4: Iterate over all the attachments and group similar files.
			similarAttachments := make(map[string][]*attachment.Attachment)
			for _, a := range attachments {

				// Extract the array of attachments that are similar
				arr := similarAttachments[a.ObjectKey]

				// Add attachment to the similar attachments.
				arr = append(arr, a)

				// Update our staff user's attachments array.
				similarAttachments[a.ObjectKey] = arr
			}
			log.Println(similarAttachments) // For debugging purposes only.

			// DEVELOPERS NOTE:
			// Once we are here then we need to (1) pick one attachment and
			// (2) delete remaining similar attachments.

			// STEP 5: Iterate through all similar files.
			for objectKey, attachmentsByObjectKey := range similarAttachments {
				log.Println("////////////////////////////////////////////////////////////////////////////////////////////////////////////")
				log.Println("////////////////////////////////////////////////////////////////////////////////////////////////////////////")
				log.Println("////////////////////////////////////////////////////////////////////////////////////////////////////////////")
				log.Println("objectKey:", objectKey)
				log.Println("------------------------------------------------------------------------------------------------------------")

				var uniqueAttachment *attachment.Attachment
				var duplicateAttachments []*attachment.Attachment

				for _, a := range attachmentsByObjectKey {
					if uniqueAttachment == nil {
						uniqueAttachment = a
					} else {
						duplicateAttachments = append(duplicateAttachments, a)
					}

				}

				// STEP 6: Iterate over duplicates and delete RECORD ONLY, DO
				// NOT DELETE FILE IN S3!

				log.Println("keeping --->", uniqueAttachment.ObjectKey, uniqueAttachment.ID)
				for _, dup := range duplicateAttachments {

					// For defensive code purposes, just do a few tests to
					// make sure the record is similar before proceeding to
					// delete.
					if uniqueAttachment.StaffID == dup.StaffID && uniqueAttachment.ObjectKey == dup.ObjectKey {
						log.Println("remove --->", dup.ObjectKey, dup.ID)
						if err := attachStorer.DeleteByID(sessCtx, dup.ID); err != nil {
							log.Println("error deleting staff attachment:", err)
							return nil, err
						}
					}
				}
			}
			//------ END
		}

		return nil, nil
	}

	// Start a transaction
	if _, err := session.WithTransaction(context.Background(), transactionFunc); err != nil {
		log.Fatal(err)
	}
	return nil
}

//
// func hotfix02Order(
// 	mc *mongo.Client,
// 	ts tenant_ds.TenantStorer,
// 	userStorer user_ds.UserStorer,
// 	orderStorer s_ds.OrderStorer,
// 	hhStorer hh_ds.HowHearAboutUsItemStorer,
// 	attachStorer attachment.AttachmentStorer,
// 	tenant *tenant_ds.Tenant,
// ) error {
// 	////
// 	//// Start the transaction.
// 	////
//
// 	session, err := mc.StartSession()
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	defer session.EndSession(context.Background())
//
// 	// Define a transaction function with a series of operations
// 	transactionFunc := func(sessCtx mongo.SessionContext) (interface{}, error) {
// 		// STEP 1: Get all the attachments in our system that belong to
// 		// orders only.
// 		res, err := attachStorer.ListByType(sessCtx, attachment.AttachmentTypeOrder)
// 		if err != nil {
// 			return nil, err
// 		}
//
// 		// STEP 2: Group all the attachments per order.
// 		orderAttachments := make(map[string][]*attachment.Attachment)
// 		for _, a := range res.Results {
// 			// // For debugging purposes only.
// 			// log.Println("ID:", a.ID)
// 			// log.Println("OrderID:", a.OrderID)
// 			// log.Println("Filename:", a.Filename)
// 			// log.Println("ObjectKey:", a.ObjectKey)
// 			// log.Println("ObjectURL:", a.ObjectURL)
// 			// log.Println()
//
// 			// Extract the array of attachments that belong to a particular
// 			// order user.
// 			arr := orderAttachments[a.OrderID.Hex()]
//
// 			// Add attachment to the order user's attachments.
// 			arr = append(arr, a)
//
// 			// Update our order user's attachments array.
// 			orderAttachments[a.OrderID.Hex()] = arr
// 		}
//
// 		// log.Println(orderAttachments) // For debugging purposes only.
//
// 		// STEP 3: Iterate through all the orders.
// 		for _, attachments := range orderAttachments {
// 			// log.Println(orderID, attachments) // For debugging purposes only.
//
// 			// STEP 4: Iterate over all the attachments and group similar files.
// 			similarAttachments := make(map[string][]*attachment.Attachment)
// 			for _, a := range attachments {
//
// 				// Extract the array of attachments that are similar
// 				arr := similarAttachments[a.ObjectKey]
//
// 				// Add attachment to the similar attachments.
// 				arr = append(arr, a)
//
// 				// Update our order user's attachments array.
// 				similarAttachments[a.ObjectKey] = arr
// 			}
// 			log.Println(similarAttachments) // For debugging purposes only.
//
// 			// DEVELOPERS NOTE:
// 			// Once we are here then we need to (1) pick one attachment and
// 			// (2) delete remaining similar attachments.
//
// 			// STEP 5: Iterate through all similar files.
// 			for objectKey, attachmentsByObjectKey := range similarAttachments {
// 				log.Println("////////////////////////////////////////////////////////////////////////////////////////////////////////////")
// 				log.Println("////////////////////////////////////////////////////////////////////////////////////////////////////////////")
// 				log.Println("////////////////////////////////////////////////////////////////////////////////////////////////////////////")
// 				log.Println("objectKey:", objectKey)
// 				log.Println("------------------------------------------------------------------------------------------------------------")
//
// 				var uniqueAttachment *attachment.Attachment
// 				var duplicateAttachments []*attachment.Attachment
//
// 				for _, a := range attachmentsByObjectKey {
// 					if uniqueAttachment == nil {
// 						uniqueAttachment = a
// 					} else {
// 						duplicateAttachments = append(duplicateAttachments, a)
// 					}
//
// 				}
//
// 				// STEP 6: Iterate over duplicates and delete RECORD ONLY, DO
// 				// NOT DELETE FILE IN S3!
//
// 				log.Println("keeping --->", uniqueAttachment.ObjectKey, uniqueAttachment.ID)
// 				for _, dup := range duplicateAttachments {
//
// 					// For defensive code purposes, just do a few tests to
// 					// make sure the record is similar before proceeding to
// 					// delete.
// 					if uniqueAttachment.OrderID == dup.OrderID && uniqueAttachment.ObjectKey == dup.ObjectKey {
// 						log.Println("remove --->", dup.ObjectKey, dup.ID)
// 						// if err := attachStorer.DeleteByID(sessCtx, dup.ID); err != nil {
// 						// 	log.Println("error deleting order attachment:", err)
// 						// 	return nil, err
// 						// }
// 					}
// 				}
// 			}
// 			//------ END
// 		}
//
// 		return nil, nil
// 	}
//
// 	// Start a transaction
// 	if _, err := session.WithTransaction(context.Background(), transactionFunc); err != nil {
// 		log.Fatal(err)
// 	}
// 	return nil
// }
