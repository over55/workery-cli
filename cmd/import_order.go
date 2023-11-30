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
	o_ds "github.com/over55/workery-cli/app/order/datastore"
	sf_ds "github.com/over55/workery-cli/app/servicefee/datastore"
	ss_ds "github.com/over55/workery-cli/app/skillset/datastore"
	t_ds "github.com/over55/workery-cli/app/tenant/datastore"
	u_ds "github.com/over55/workery-cli/app/user/datastore"
	"github.com/over55/workery-cli/config"
)

func init() {
	rootCmd.AddCommand(importOrderCmd)
}

var importOrderCmd = &cobra.Command{
	Use:   "import_order",
	Short: "Import the orders from old database",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		cfg := config.New()
		mc := mongodb.NewStorage(cfg)
		ppc := postgres.NewStorage(cfg, cfg.PostgresDB.DatabasePublicSchemaName)
		lpc := postgres.NewStorage(cfg, cfg.PostgresDB.DatabaseLondonSchemaName)

		defaultLogger := slog.Default()

		tStorer := t_ds.NewDatastore(cfg, defaultLogger, mc)
		oStorer := o_ds.NewDatastore(cfg, defaultLogger, mc)
		uStorer := u_ds.NewDatastore(cfg, defaultLogger, mc)
		aStorer := a_ds.NewDatastore(cfg, defaultLogger, mc)
		cStorer := c_ds.NewDatastore(cfg, defaultLogger, mc)
		ssStorer := ss_ds.NewDatastore(cfg, defaultLogger, mc)
		sfStorer := sf_ds.NewDatastore(cfg, defaultLogger, mc)

		tenant, err := tStorer.GetBySchemaName(ctx, cfg.PostgresDB.DatabaseLondonSchemaName)
		if err != nil {
			log.Fatal(err)
		}

		RunImportOrder(cfg, ppc, lpc, oStorer, uStorer, aStorer, cStorer, ssStorer, sfStorer, tenant)
	},
}

func RunImportOrder(
	cfg *config.Conf,
	public *sql.DB,
	london *sql.DB,
	oStorer o_ds.OrderStorer,
	uStorer u_ds.UserStorer,
	aStorer a_ds.AssociateStorer,
	cStorer c_ds.CustomerStorer,
	ssStorer ss_ds.SkillSetStorer,
	sfStorer sf_ds.ServiceFeeStorer,
	tenant *t_ds.Tenant,
) {
	fmt.Println("Beginning importing orders")
	data, err := ListAllWorkOrders(london)
	if err != nil {
		log.Fatal(err)
	}

	for _, datum := range data {
		importOrder(context.Background(), oStorer, uStorer, aStorer, cStorer, ssStorer, sfStorer, tenant, datum)
	}
	fmt.Println("Finished importing orders")
}

type OldWorkOrder struct {
	ID                                        uint64      `json:"id"`
	AssociateID                               null.Int    `json:"associate_id"`
	CustomerID                                uint64      `json:"customer_id"`
	Description                               string      `json:"description"`
	AssignmentDate                            null.Time   `json:"assignment_date"`
	IsOngoing                                 bool        `json:"is_ongoing"`
	IsHomeSupportService                      bool        `json:"is_home_support_service"`
	StartDate                                 time.Time   `json:"start_date"`
	CompletionDate                            null.Time   `json:"completion_date"`
	Hours                                     float64     `json:"hours"`
	TypeOf                                    int8        `json:"type_of"`
	IndexedText                               string      `json:"indexed_text"`
	ClosingReason                             int8        `json:"closing_reason"`
	ClosingReasonOther                        null.String `json:"closing_reason_other"`
	State                                     string      `json:"state"`
	WasJobSatisfactory                        bool        `json:"was_job_satisfactory"`
	WasJobFinishedOnTimeAndOnBudget           bool        `json:"was_job_finished_on_time_and_on_budget"`
	WasAssociatePunctual                      bool        `json:"was_associate_punctual"`
	WasAssociateProfessional                  bool        `json:"was_associate_professional"`
	WouldCustomerReferOurOrganization         bool        `json:"would_customer_refer_our_organization"`
	Score                                     float64     `json:"score"`
	InvoiceDate                               null.Time   `json:"invoice_date"`
	InvoiceQuoteAmountCurrency                string      `json:"invoice_quote_amount_currency"`
	InvoiceQuoteAmount                        float64     `json:"invoice_quote_amount"`
	InvoiceLabourAmountCurrency               string      `json:"invoice_labour_amount_currency"`
	InvoiceLabourAmount                       float64     `json:"invoice_labour_amount"`
	InvoiceMaterialAmountCurrency             string      `json:"invoice_material_amount_currency"`
	InvoiceMaterialAmount                     float64     `json:"invoice_material_amount"`
	InvoiceTaxAmountCurrency                  string      `json:"invoice_tax_amount_currency"`
	InvoiceTaxAmount                          float64     `json:"invoice_tax_amount"`
	InvoiceTotalAmountCurrency                string      `json:"invoice_total_amount_currency"`
	InvoiceTotalAmount                        float64     `json:"invoice_total_amount"`
	InvoiceServiceFeeAmountCurrency           string      `json:"invoice_service_fee_amount_currency"`
	InvoiceServiceFeeAmount                   float64     `json:"invoice_service_fee_amount"`
	InvoiceServiceFeePaymentDate              null.Time   `json:"invoice_service_fee_payment_date"`
	Created                                   time.Time   `json:"created"`
	CreatedByID                               null.Int    `json:"created_by_id"`
	CreatedFrom                               null.String `json:"created_from"`
	LastModified                              time.Time   `json:"last_modified"`
	LastModifiedByID                          null.Int    `json:"last_modified_by_id"`
	LastModifiedFrom                          null.String `json:"last_modified_from"`
	InvoiceServiceFeeID                       null.Int    `json:"invoice_service_fee_id"`
	LatestPendingTaskID                       null.Int    `json:"latest_pending_task_id"`
	OngoingWorkOrderID                        null.Int    `json:"ongoing_work_order_id"`
	WasSurveyConducted                        bool        `json:"was_survey_conducted"`
	WasThereFinancialsInputted                bool        `json:"was_there_financials_inputted"`
	InvoiceActualServiceFeeAmountPaidCurrency string      `json:"invoice_actual_service_fee_amount_paid_currency"`
	InvoiceActualServiceFeeAmountPaid         float64     `json:"invoice_actual_service_fee_amount_paid"`
	InvoiceBalanceOwingAmountCurrency         string      `json:"invoice_balance_owing_amount_currency"`
	InvoiceBalanceOwingAmount                 float64     `json:"invoice_balance_owing_amount"`
	InvoiceQuotedLabourAmountCurrency         string      `json:"invoice_quoted_labour_amount_currency"`
	InvoiceQuotedLabourAmount                 float64     `json:"invoice_quoted_labour_amount"`
	InvoiceQuotedMaterialAmountCurrency       string      `json:"invoice_quoted_material_amount_currency"`
	InvoiceQuotedMaterialAmount               float64     `json:"invoice_quoted_material_amount"`
	InvoiceTotalQuoteAmountCurrency           string      `json:"invoice_total_quote_amount_currency"`
	InvoiceTotalQuoteAmount                   float64     `json:"invoice_total_quote_amount"`
	Visits                                    int8        `json:"visits"`
	InvoiceIDs                                null.String `json:"invoice_ids"`
	NoSurveyConductedReason                   null.Int    `json:"no_survey_conducted_reason"`
	NoSurveyConductedReasonOther              null.String `json:"no_survey_conducted_reason_other"`
	ClonedFromID                              null.Int    `json:"cloned_from_id"`
	InvoiceDepositAmountCurrency              string      `json:"invoice_deposit_amount_currency"`
	InvoiceDepositAmount                      float64     `json:"invoice_deposit_amount"`
	InvoiceOtherCostsAmountCurrency           string      `json:"invoice_other_costs_amount_currency"`
	InvoiceOtherCostsAmount                   float64     `json:"invoice_other_costs_amount"`
	InvoiceQuotedOtherCostsAmountCurrency     string      `json:"invoice_quoted_other_costs_amount_currency"`
	InvoiceQuotedOtherCostsAmount             float64     `json:"invoice_quoted_other_costs_amount"`
	InvoicePaidTo                             null.Int    `json:"invoice_paid_to"`
	InvoiceAmountDueCurrency                  string      `json:"invoice_amount_due_currency"`
	InvoiceAmountDue                          float64     `json:"invoice_amount_due"`
	InvoiceSubTotalAmountCurrency             string      `json:"invoice_sub_total_amount_currency"`
	InvoiceSubTotalAmount                     float64     `json:"invoice_sub_total_amount"`
	ClosingReasonComment                      string      `json:"closing_reason_comment"`
}

func ListAllWorkOrders(db *sql.DB) ([]*OldWorkOrder, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `
	SELECT
        id, associate_id, customer_id, description, assignment_date, is_ongoing, is_home_support_service, start_date,
		completion_date, hours, type_of, indexed_text, closing_reason, closing_reason_other, state,
		was_job_satisfactory, was_job_finished_on_time_and_on_budget, was_associate_punctual, was_associate_professional, would_customer_refer_our_organization,
		score, invoice_date, invoice_quote_amount_currency, invoice_quote_amount, invoice_labour_amount_currency, invoice_labour_amount,
		invoice_material_amount_currency, invoice_material_amount, invoice_tax_amount_currency, invoice_tax_amount,
		invoice_total_amount_currency, invoice_total_amount, invoice_service_fee_amount_currency, invoice_service_fee_amount, invoice_service_fee_payment_date,
		created, created_by_id, created_from, last_modified, last_modified_by_id, last_modified_from, invoice_service_fee_id, latest_pending_task_id, ongoing_work_order_id,
		was_survey_conducted, was_there_financials_inputted, invoice_actual_service_fee_amount_paid_currency, invoice_actual_service_fee_amount_paid,
		invoice_balance_owing_amount_currency, invoice_balance_owing_amount, invoice_quoted_labour_amount_currency, invoice_quoted_labour_amount,
		invoice_quoted_material_amount_currency, invoice_quoted_material_amount, invoice_total_quote_amount_currency, invoice_total_quote_amount, visits, invoice_ids,
		no_survey_conducted_reason, no_survey_conducted_reason_other, cloned_from_id, invoice_deposit_amount_currency, invoice_deposit_amount,
		invoice_other_costs_amount_currency, invoice_other_costs_amount, invoice_quoted_other_costs_amount_currency, invoice_quoted_other_costs_amount, invoice_paid_to,
		invoice_amount_due_currency, invoice_amount_due, invoice_sub_total_amount_currency, invoice_sub_total_amount, closing_reason_comment
	FROM
        london.workery_work_orders
	ORDER BY
	    id
	ASC
	`
	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}

	var arr []*OldWorkOrder
	defer rows.Close()
	for rows.Next() {
		m := new(OldWorkOrder)
		err = rows.Scan(
			&m.ID, &m.AssociateID, &m.CustomerID, &m.Description, &m.AssignmentDate, &m.IsOngoing, &m.IsHomeSupportService, &m.StartDate,
			&m.CompletionDate, &m.Hours, &m.TypeOf, &m.IndexedText, &m.ClosingReason, &m.ClosingReasonOther, &m.State,
			&m.WasJobSatisfactory, &m.WasJobFinishedOnTimeAndOnBudget, &m.WasAssociatePunctual, &m.WasAssociateProfessional, &m.WouldCustomerReferOurOrganization,
			&m.Score, &m.InvoiceDate, &m.InvoiceQuoteAmountCurrency, &m.InvoiceQuoteAmount, &m.InvoiceLabourAmountCurrency, &m.InvoiceLabourAmount,
			&m.InvoiceMaterialAmountCurrency, &m.InvoiceMaterialAmount, &m.InvoiceTaxAmountCurrency, &m.InvoiceTaxAmount,
			&m.InvoiceTotalAmountCurrency, &m.InvoiceTotalAmount, &m.InvoiceServiceFeeAmountCurrency, &m.InvoiceServiceFeeAmount, &m.InvoiceServiceFeePaymentDate,
			&m.Created, &m.CreatedByID, &m.CreatedFrom, &m.LastModified, &m.LastModifiedByID, &m.LastModifiedFrom, &m.InvoiceServiceFeeID, &m.LatestPendingTaskID, &m.OngoingWorkOrderID,
			&m.WasSurveyConducted, &m.WasThereFinancialsInputted, &m.InvoiceActualServiceFeeAmountPaidCurrency, &m.InvoiceActualServiceFeeAmountPaid,
			&m.InvoiceBalanceOwingAmountCurrency, &m.InvoiceBalanceOwingAmount, &m.InvoiceQuotedLabourAmountCurrency, &m.InvoiceQuotedLabourAmount,
			&m.InvoiceQuotedMaterialAmountCurrency, &m.InvoiceQuotedMaterialAmount, &m.InvoiceTotalQuoteAmountCurrency, &m.InvoiceTotalQuoteAmount, &m.Visits, &m.InvoiceIDs,
			&m.NoSurveyConductedReason, &m.NoSurveyConductedReasonOther, &m.ClonedFromID, &m.InvoiceDepositAmountCurrency, &m.InvoiceDepositAmount,
			&m.InvoiceOtherCostsAmountCurrency, &m.InvoiceOtherCostsAmount, &m.InvoiceQuotedOtherCostsAmountCurrency, &m.InvoiceQuotedOtherCostsAmount, &m.InvoicePaidTo,
			&m.InvoiceAmountDueCurrency, &m.InvoiceAmountDue, &m.InvoiceSubTotalAmountCurrency, &m.InvoiceSubTotalAmount, &m.ClosingReasonComment,
		)
		if err != nil {
			log.Fatal("ListAllWorkOrders | rows.Scan", err)
		}
		arr = append(arr, m)
	}
	err = rows.Err()
	if err != nil {
		log.Fatal("ListAllWorkOrders | rows.Err", err)
	}
	return arr, err
}

func importOrder(
	ctx context.Context,
	oStorer o_ds.OrderStorer,
	uStorer u_ds.UserStorer,
	aStorer a_ds.AssociateStorer,
	cStorer c_ds.CustomerStorer,
	ssStorer ss_ds.SkillSetStorer,
	sfStorer sf_ds.ServiceFeeStorer,
	tenant *t_ds.Tenant,
	wo *OldWorkOrder,
) {
	//
	// Get the optional `Associate` data to compile `name`, `lexical name`, 'gender', and 'birthdate' field.
	//

	var associateID primitive.ObjectID = primitive.NilObjectID
	var associateName string
	var associateLexicalName string
	var associateGender int8
	var associateGenderOther string
	var associateBirthdate time.Time
	var associateEmail string
	var associatePhone string
	var associatePhoneType int8
	var associatePhoneExtension string
	var associateOtherPhone string
	var associateOtherPhoneType int8
	var associateOtherPhoneExtension string
	var associateFullAddressWithoutPostalCode string
	var associateFullAddressURL string
	a, err := aStorer.GetByOldID(ctx, uint64(wo.AssociateID.ValueOrZero()))
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
		associateFullAddressWithoutPostalCode = a.FullAddressWithoutPostalCode
		associateFullAddressURL = a.FullAddressURL
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
	var customerEmail string
	var customerPhone string
	var customerPhoneType int8
	var customerPhoneExtension string
	var customerOtherPhone string
	var customerOtherPhoneType int8
	var customerOtherPhoneExtension string
	var customerFullAddressWithoutPostalCode string
	var customerFullAddressURL string
	c, err := cStorer.GetByOldID(ctx, wo.CustomerID)
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
		customerFullAddressWithoutPostalCode = c.FullAddressWithoutPostalCode
		customerFullAddressURL = c.FullAddressURL
	}

	//
	// Compile our `state`.
	//

	var state int8
	switch s := wo.State; s {
	case "new":
		state = o_ds.OrderNewState
	case "declined":
		state = o_ds.OrderDeclinedState
	case "pending":
		state = o_ds.OrderPendingState
	case "cancelled":
		state = o_ds.OrderCancelledState
	case "ongoing":
		state = o_ds.OrderOngoingState
	case "in_progress":
		state = o_ds.OrderInProgressState
	case "completed_and_unpaid":
		state = o_ds.OrderCompletedButUnpaidState
	case "completed_but_unpaid":
		state = o_ds.OrderCompletedButUnpaidState
	case "completed_and_paid":
		state = o_ds.OrderCompletedAndPaidState
	case "archived":
		state = o_ds.OrderArchivedState
	default:
		state = o_ds.OrderArchivedState
	}

	//
	// Compile `createdById` and `createdByName` values.
	//

	var createdByUserID primitive.ObjectID = primitive.NilObjectID
	var createdByUserName string
	createdByUser, err := uStorer.GetByOldID(ctx, uint64(wo.CreatedByID.ValueOrZero()))
	if err != nil {
		log.Fatal(err)
	}
	if createdByUser != nil {
		createdByUserID = createdByUser.ID
		createdByUserName = createdByUser.Name
	}

	//
	// Get `modifiedByID` and `modifiedByName` values.
	//

	var modifiedByUserID primitive.ObjectID = primitive.NilObjectID
	var modifiedByUserName string
	modifiedByUser, err := uStorer.GetByOldID(ctx, uint64(wo.LastModifiedByID.ValueOrZero()))
	if err != nil {
		log.Fatal(err)
	}
	if modifiedByUser != nil {
		modifiedByUserID = modifiedByUser.ID
		modifiedByUserName = modifiedByUser.Name
	}

	//
	// Get invoice service fee.
	//

	var invoiceServiceFeeID primitive.ObjectID = primitive.NilObjectID
	var invoiceServiceFeeName string
	var invoiceServiceFeeDescription string
	var invoiceServiceFeePercentage float64
	sf, err := sfStorer.GetByOldID(ctx, uint64(wo.InvoiceServiceFeeID.ValueOrZero()))
	if err != nil {
		log.Fatal(err)
	}
	if sf != nil {
		invoiceServiceFeeID = sf.ID
		invoiceServiceFeeName = sf.Name
		invoiceServiceFeeDescription = sf.Description
		invoiceServiceFeePercentage = sf.Percentage
	}

	//
	// Get clonedFromOrderID
	//

	var clonedFromOrderID primitive.ObjectID = primitive.NilObjectID
	clonedOrder, err := oStorer.GetByWJID(ctx, uint64(wo.ClonedFromID.ValueOrZero()))
	if err != nil {
		log.Fatal(err)
	}
	if clonedOrder != nil {
		clonedFromOrderID = clonedOrder.ID
	}

	//
	// Set empty order tags.
	//

	var orderTags = make([]*o_ds.OrderTag, 0)

	//
	// Set empty order skill sets.
	//

	var orderSkillSets = make([]*o_ds.OrderSkillSet, 0)

	//
	// Insert the record
	//

	m := &o_ds.Order{
		WJID:                                  wo.ID,
		TenantID:                              tenant.ID,
		TenantIDWithWJID:                      fmt.Sprintf("%v_%v", tenant.ID.Hex(), wo.ID),
		ID:                                    primitive.NewObjectID(),
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
		Description:                           wo.Description,
		AssignmentDate:                        wo.AssignmentDate.ValueOrZero(),
		IsOngoing:                             wo.IsOngoing,
		IsHomeSupportService:                  wo.IsHomeSupportService,
		StartDate:                             wo.StartDate,
		CompletionDate:                        wo.CompletionDate.ValueOrZero(),
		Hours:                                 wo.Hours,
		Type:                                  wo.TypeOf,
		IndexedText:                           wo.IndexedText,
		ClosingReason:                         wo.ClosingReason,
		ClosingReasonOther:                    wo.ClosingReasonOther.ValueOrZero(),
		Status:                                state,
		Currency:                              "CAD",
		WasJobSatisfactory:                    wo.WasJobSatisfactory,
		WasJobFinishedOnTimeAndOnBudget:       wo.WasJobFinishedOnTimeAndOnBudget,
		WasAssociatePunctual:                  wo.WasAssociatePunctual,
		WasAssociateProfessional:              wo.WasAssociateProfessional,
		WouldCustomerReferOurOrganization:     wo.WouldCustomerReferOurOrganization,
		Score:                                 wo.Score,
		InvoiceDate:                           wo.InvoiceDate.ValueOrZero(),
		InvoiceQuoteAmount:                    wo.InvoiceQuoteAmount,
		InvoiceLabourAmount:                   wo.InvoiceLabourAmount,
		InvoiceMaterialAmount:                 wo.InvoiceMaterialAmount,
		InvoiceTaxAmount:                      wo.InvoiceTaxAmount,
		InvoiceTotalAmount:                    wo.InvoiceTotalAmount,
		InvoiceServiceFeeAmount:               wo.InvoiceServiceFeeAmount,
		InvoiceServiceFeePaymentDate:          wo.InvoiceServiceFeePaymentDate.ValueOrZero(),
		CreatedAt:                             wo.Created,
		CreatedByUserID:                       createdByUserID,
		CreatedByUserName:                     createdByUserName,
		CreatedFromIPAddress:                  wo.CreatedFrom.ValueOrZero(),
		ModifiedAt:                            wo.LastModified,
		ModifiedByUserID:                      modifiedByUserID,
		ModifiedByUserName:                    modifiedByUserName,
		ModifiedFromIPAddress:                 wo.LastModifiedFrom.ValueOrZero(),
		InvoiceServiceFeeID:                   invoiceServiceFeeID,
		InvoiceServiceFeeName:                 invoiceServiceFeeName,
		InvoiceServiceFeeDescription:          invoiceServiceFeeDescription,
		InvoiceServiceFeePercentage:           invoiceServiceFeePercentage,
		// OngoingWorkOrderID:                ongoingWorkOrderID,
		WasSurveyConducted:                wo.WasSurveyConducted,
		WasThereFinancialsInputted:        wo.WasThereFinancialsInputted,
		InvoiceActualServiceFeeAmountPaid: wo.InvoiceActualServiceFeeAmountPaid,
		InvoiceBalanceOwingAmount:         wo.InvoiceBalanceOwingAmount,
		InvoiceQuotedLabourAmount:         wo.InvoiceQuotedLabourAmount,
		InvoiceQuotedMaterialAmount:       wo.InvoiceQuotedMaterialAmount,
		InvoiceTotalQuoteAmount:           wo.InvoiceTotalQuoteAmount,
		Visits:                            wo.Visits,
		InvoiceIDs:                        wo.InvoiceIDs.ValueOrZero(),
		NoSurveyConductedReason:           int8(wo.NoSurveyConductedReason.ValueOrZero()),
		NoSurveyConductedReasonOther:      wo.NoSurveyConductedReasonOther.ValueOrZero(),
		ClonedFromOrderID:                 clonedFromOrderID,
		InvoiceDepositAmount:              wo.InvoiceDepositAmount,
		InvoiceOtherCostsAmount:           wo.InvoiceOtherCostsAmount,
		InvoiceQuotedOtherCostsAmount:     wo.InvoiceQuotedOtherCostsAmount,
		InvoicePaidTo:                     int8(wo.InvoicePaidTo.ValueOrZero()),
		InvoiceAmountDue:                  wo.InvoiceAmountDue,
		InvoiceSubTotalAmount:             wo.InvoiceSubTotalAmount,
		ClosingReasonComment:              wo.ClosingReasonComment,
		Tags:                              orderTags,
		SkillSets:                         orderSkillSets,
		// LatestPendingTaskID:               wo.LatestPendingTaskID, //TODO: LATER
	}

	if err := oStorer.Create(ctx, m); err != nil {
		log.Panic(err)
	}
	fmt.Println("Imported Order ID#", m.ID)
}
