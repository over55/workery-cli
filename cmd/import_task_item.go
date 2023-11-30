package cmd

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/spf13/cobra"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/exp/slog"
	"gopkg.in/guregu/null.v4"

	"github.com/over55/workery-cli/adapter/storage/mongodb"
	"github.com/over55/workery-cli/adapter/storage/postgres"
	a_ds "github.com/over55/workery-cli/app/associate/datastore"
	c_ds "github.com/over55/workery-cli/app/customer/datastore"
	order_ds "github.com/over55/workery-cli/app/order/datastore"
	ti_ds "github.com/over55/workery-cli/app/taskitem/datastore"
	tenant_ds "github.com/over55/workery-cli/app/tenant/datastore"
	user_ds "github.com/over55/workery-cli/app/user/datastore"
	"github.com/over55/workery-cli/config"
)

func init() {
	rootCmd.AddCommand(importTaskItemCmd)
}

var importTaskItemCmd = &cobra.Command{
	Use:   "import_task_item",
	Short: "Import the tags from old database",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		cfg := config.New()
		mc := mongodb.NewStorage(cfg)
		ppc := postgres.NewStorage(cfg, cfg.PostgresDB.DatabasePublicSchemaName)
		lpc := postgres.NewStorage(cfg, cfg.PostgresDB.DatabaseLondonSchemaName)

		defaultLogger := slog.Default()

		tenantStorer := tenant_ds.NewDatastore(cfg, defaultLogger, mc)
		oStorer := order_ds.NewDatastore(cfg, defaultLogger, mc)
		tiStorer := ti_ds.NewDatastore(cfg, defaultLogger, mc)
		uStorer := user_ds.NewDatastore(cfg, defaultLogger, mc)
		aStorer := a_ds.NewDatastore(cfg, defaultLogger, mc)
		cStorer := c_ds.NewDatastore(cfg, defaultLogger, mc)

		tenant, err := tenantStorer.GetBySchemaName(ctx, cfg.PostgresDB.DatabaseLondonSchemaName)
		if err != nil {
			log.Fatal(err)
		}

		RunImportTaskItem(cfg, ppc, lpc, uStorer, oStorer, tiStorer, aStorer, cStorer, tenant)
	},
}

func RunImportTaskItem(
	cfg *config.Conf,
	public *sql.DB,
	london *sql.DB,
	uStorer user_ds.UserStorer,
	oStorer order_ds.OrderStorer,
	tiStorer ti_ds.TaskItemStorer,
	aStorer a_ds.AssociateStorer,
	cStorer c_ds.CustomerStorer,
	tenant *tenant_ds.Tenant,
) {
	fmt.Println("Beginning importing task items")
	data, err := ListAllTaskItems(london)
	if err != nil {
		log.Fatal(err)
	}

	for _, datum := range data {
		importTaskItem(context.Background(), uStorer, oStorer, tiStorer, aStorer, cStorer, tenant, datum)
	}
	fmt.Println("Finished importing task items")
}

type OldUTaskItem struct {
	ID                       uint64      `json:"id"`
	TypeOf                   int8        `json:"type_of"`
	Title                    string      `json:"title"`
	Description              string      `json:"description"`
	DueDate                  time.Time   `json:"due_date"`
	IsClosed                 bool        `json:"is_closed"`
	WasPostponed             bool        `json:"was_postponed"`
	ClosingReason            int8        `json:"closing_reason"`
	ClosingReasonOther       string      `json:"closing_reason_other"`
	CreatedAt                time.Time   `json:"created_at"`
	CreatedFrom              null.String `json:"created_from"`
	CreatedFromIsPublic      null.Bool   `json:"created_from_is_public"`
	CreatedByID              null.Int    `json:"created_by_id"`
	LastModifiedAt           time.Time   `json:"last_modified_at"`
	LastModifiedFrom         null.String `json:"last_modified_from"`
	LastModifiedFromIsPublic null.Bool   `json:"last_modified_from_is_public"`
	JobID                    uint64      `json:"job_id"`
	LastModifiedByID         null.Int    `json:"last_modified_by_id"`
	OngoingJobID             null.Int    `json:"ongoing_job_id"`
}

func ListAllTaskItems(db *sql.DB) ([]*OldUTaskItem, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `
	SELECT
	    id, type_of, title, description, due_date, is_closed, was_postponed,
		closing_reason, closing_reason_other, created_at, created_from,
		created_from_is_public, last_modified_at, last_modified_from,
		last_modified_from_is_public, created_by_id, job_id, last_modified_by_id,
		ongoing_job_id
	FROM
	    workery_task_items
	ORDER BY
	    id
	ASC
	`
	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}

	var arr []*OldUTaskItem
	defer rows.Close()
	for rows.Next() {
		m := new(OldUTaskItem)
		err = rows.Scan(
			&m.ID,
			&m.TypeOf,
			&m.Title,
			&m.Description,
			&m.DueDate,
			&m.IsClosed,
			&m.WasPostponed,
			&m.ClosingReason,
			&m.ClosingReasonOther,
			&m.CreatedAt,
			&m.CreatedFrom,
			&m.CreatedFromIsPublic,
			&m.LastModifiedAt,
			&m.LastModifiedFrom,
			&m.LastModifiedFromIsPublic,
			&m.CreatedByID,
			&m.JobID,
			&m.LastModifiedByID,
			&m.OngoingJobID,
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

func importTaskItem(
	ctx context.Context,
	uStorer user_ds.UserStorer,
	oStorer order_ds.OrderStorer,
	tiStorer ti_ds.TaskItemStorer,
	aStorer a_ds.AssociateStorer,
	cStorer c_ds.CustomerStorer,
	tenant *tenant_ds.Tenant,
	ti *OldUTaskItem,
) {
	//
	// Set the `state`.
	//

	var state int8 = 1

	//
	// Get our `OrderId` value.
	//

	order, err := oStorer.GetByWJID(ctx, ti.JobID)
	if err != nil {
		log.Fatal(err)
	}
	if order == nil {
		log.Fatal("order does not exist")
	}

	var orderSkillSets []*ti_ds.TaskItemSkillSet
	for _, ss := range order.SkillSets {
		orderSkillSets = append(orderSkillSets, &ti_ds.TaskItemSkillSet{
			ID:                    ss.ID,
			OrderID:               ss.OrderID,
			OrderWJID:             ss.OrderWJID,
			OrderTenantIDWithWJID: ss.OrderTenantIDWithWJID,
			TenantID:              ss.TenantID,
			Category:              ss.Category,
			SubCategory:           ss.SubCategory,
			Description:           ss.Description,
			OldID:                 ss.OldID,
		})
	}

	var orderTags []*ti_ds.TaskItemOrderTag
	for _, t := range order.Tags {
		orderTags = append(orderTags, &ti_ds.TaskItemOrderTag{
			ID:          t.ID,
			TenantID:    t.TenantID,
			Text:        t.Text,
			Description: t.Description,
			OldID:       t.OldID,
		})
	}

	//
	// Get created by
	//

	var createdByUserID primitive.ObjectID = primitive.NilObjectID
	var createdByUserName string
	createdByUser, _ := uStorer.GetByOldID(ctx, uint64(ti.CreatedByID.ValueOrZero()))
	if createdByUser != nil {
		createdByUserID = createdByUser.ID
		createdByUserName = createdByUser.Name
	}

	//
	// Get modified by
	//

	var modifiedByUserID primitive.ObjectID = primitive.NilObjectID
	var modifiedByUserName string
	modifiedByUser, _ := uStorer.GetByOldID(ctx, uint64(ti.CreatedByID.ValueOrZero()))
	if modifiedByUser != nil {
		modifiedByUserID = modifiedByUser.ID
		modifiedByUserName = modifiedByUser.Name
	}

	//
	// Get the optional `Associate` data to compile `name`, `lexical name`, 'gender', and 'birthdate' field.
	//

	var associateID primitive.ObjectID = primitive.NilObjectID
	var associateName string
	var associateLexicalName string
	var associateGender int8
	var associateGenderOther string
	var associateBirthdate time.Time
	var associateTags = make([]*ti_ds.TaskItemAssociateTag, 0)
	var associateEmail string
	var associatePhone string
	var associatePhoneType int8
	var associatePhoneExtension string
	var associateOtherPhone string
	var associateOtherPhoneType int8
	var associateOtherPhoneExtension string
	var associateFullAddressWithoutPostalCode string
	var associateFullAddressURL string

	a, err := aStorer.GetByID(ctx, order.AssociateID)
	if err != nil {
		log.Fatal(err)
	}
	if a != nil {
		associateID = a.ID
		associateName = a.Name
		associateLexicalName = a.LexicalName
		associateGender = a.Gender
		associateGenderOther = a.GenderOther
		associateBirthdate = a.BirthDate
		associateEmail = a.Email
		associatePhone = a.Phone
		associatePhoneType = a.PhoneType
		associatePhoneExtension = a.PhoneExtension
		associateOtherPhone = a.OtherPhone
		associateOtherPhoneType = a.OtherPhoneType
		associateOtherPhoneExtension = a.OtherPhoneExtension
		associateFullAddressWithoutPostalCode = a.FullAddressWithoutPostalCode
		associateFullAddressURL = a.FullAddressURL

		for _, tag := range a.Tags {
			associateTags = append(associateTags, &ti_ds.TaskItemAssociateTag{
				ID:          tag.ID,
				TenantID:    tag.TenantID,
				Text:        tag.Text,
				Description: tag.Description,
				OldID:       tag.OldID,
			})
		}
	}

	//
	// Generate our full name / lexical full name / gender / DOB.
	//

	var customerID primitive.ObjectID = primitive.NilObjectID
	var customerName string
	var customerLexicalName string
	var customerGender int8
	var customerGenderOther string
	var customerDOB time.Time
	var customerTags []*ti_ds.TaskItemCustomerTag
	var customerEmail string
	var customerPhone string
	var customerPhoneType int8
	var customerPhoneExtension string
	var customerOtherPhone string
	var customerOtherPhoneType int8
	var customerOtherPhoneExtension string
	var customerFullAddressWithoutPostalCode string
	var customerFullAddressURL string

	c, err := cStorer.GetByID(ctx, order.CustomerID)
	if err != nil {
		log.Fatal(err)
	}
	if c != nil {
		customerID = c.ID
		customerName = c.Name
		customerLexicalName = c.LexicalName
		customerGender = c.Gender
		customerGenderOther = c.GenderOther
		customerDOB = c.BirthDate
		customerEmail = c.Email
		customerPhone = c.Phone
		customerPhoneType = c.PhoneType
		customerPhoneExtension = c.PhoneExtension
		customerOtherPhone = c.OtherPhone
		customerOtherPhoneType = c.OtherPhoneType
		customerOtherPhoneExtension = c.OtherPhoneExtension
		customerFullAddressWithoutPostalCode = c.FullAddressWithoutPostalCode
		customerFullAddressURL = c.FullAddressURL

		for _, tag := range c.Tags {
			customerTags = append(customerTags, &ti_ds.TaskItemCustomerTag{
				ID:          tag.ID,
				TenantID:    tag.TenantID,
				Text:        tag.Text,
				Description: tag.Description,
				OldID:       tag.OldID,
			})
		}
	}

	//
	// Insert create task item.
	//

	m := &ti_ds.TaskItem{
		ID:                                    primitive.NewObjectID(),
		Type:                                  ti.TypeOf,
		Title:                                 ti.Title,
		Description:                           ti.Description,
		DueDate:                               ti.DueDate,
		IsClosed:                              ti.IsClosed,
		WasPostponed:                          ti.WasPostponed,
		ClosingReason:                         ti.ClosingReason,
		ClosingReasonOther:                    ti.ClosingReasonOther,
		OrderID:                               order.ID,
		OrderWJID:                             order.WJID,
		OrderTenantIDWithWJID:                 fmt.Sprintf("%v_%v", order.TenantID.Hex(), order.WJID),
		OrderStartDate:                        order.StartDate,
		OrderDescription:                      order.Description,
		CreatedAt:                             ti.CreatedAt,
		CreatedByUserID:                       createdByUserID,
		CreatedByUserName:                     createdByUserName,
		CreatedFromIPAddress:                  ti.CreatedFrom.ValueOrZero(),
		ModifiedByUserID:                      modifiedByUserID,
		ModifiedByUserName:                    modifiedByUserName,
		ModifiedFromIPAddress:                 ti.LastModifiedFrom.ValueOrZero(),
		Status:                                state,
		TenantID:                              tenant.ID,
		OldID:                                 ti.ID,
		AssociateID:                           associateID,
		AssociateName:                         associateName,
		AssociateLexicalName:                  associateLexicalName,
		AssociateGender:                       associateGender,
		AssociateGenderOther:                  associateGenderOther,
		AssociateBirthdate:                    associateBirthdate,
		AssociateEmail:                        associateEmail,
		AssociatePhone:                        associatePhone,
		AssociatePhoneType:                    associatePhoneType,
		AssociatePhoneExtension:               associatePhoneExtension,
		AssociateOtherPhone:                   associateOtherPhone,
		AssociateOtherPhoneType:               associateOtherPhoneType,
		AssociateOtherPhoneExtension:          associateOtherPhoneExtension,
		AssociateFullAddressWithoutPostalCode: associateFullAddressWithoutPostalCode,
		AssociateFullAddressURL:               associateFullAddressURL,
		CustomerID:                            customerID,
		CustomerName:                          customerName,
		CustomerLexicalName:                   customerLexicalName,
		CustomerGender:                        customerGender,
		CustomerGenderOther:                   customerGenderOther,
		CustomerBirthdate:                     customerDOB,
		CustomerEmail:                         customerEmail,
		CustomerPhone:                         customerPhone,
		CustomerPhoneType:                     customerPhoneType,
		CustomerPhoneExtension:                customerPhoneExtension,
		CustomerOtherPhone:                    customerOtherPhone,
		CustomerOtherPhoneType:                customerOtherPhoneType,
		CustomerOtherPhoneExtension:           customerOtherPhoneExtension,
		CustomerFullAddressWithoutPostalCode:  customerFullAddressWithoutPostalCode,
		CustomerFullAddressURL:                customerFullAddressURL,
		CustomerTags:                          customerTags,
		AssociateTags:                         associateTags,
		OrderSkillSets:                        orderSkillSets,
		OrderTags:                             orderTags,
	}

	if err := tiStorer.Create(ctx, m); err != nil {
		log.Panic(err)
	}

	order.LatestPendingTaskID = m.ID
	order.LatestPendingTaskTitle = m.Title
	order.LatestPendingTaskDescription = m.Description
	order.LatestPendingTaskDueDate = m.DueDate
	order.LatestPendingTaskType = m.Type

	if err := oStorer.UpdateByID(ctx, order); err != nil {
		log.Panic(err)
	}

	fmt.Println("Imported TaskItem ID#", m.ID.Hex(), " and updated Order ID#", order.ID.Hex())
}
