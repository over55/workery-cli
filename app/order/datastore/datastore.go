package datastore

import (
	"context"
	"log"
	"log/slog"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	c "github.com/over55/workery-cli/config"
)

type Order struct {
	ID                                    primitive.ObjectID           `bson:"_id" json:"id"`
	WJID                                  uint64                       `bson:"wjid" json:"wjid"` // A.K.A. `Workery Job ID` (Not including tenancy)
	CustomerID                            primitive.ObjectID           `bson:"customer_id" json:"customer_id"`
	CustomerFirstName                     string                       `bson:"customer_first_name" json:"customer_first_name,omitempty"`
	CustomerLastName                      string                       `bson:"customer_last_name" json:"customer_last_name,omitempty"`
	CustomerName                          string                       `bson:"customer_name" json:"customer_name,omitempty"`
	CustomerLexicalName                   string                       `bson:"customer_lexical_name" json:"customer_lexical_name,omitempty"`
	CustomerGender                        int8                         `bson:"customer_gender" json:"customer_gender"`
	CustomerGenderOther                   string                       `bson:"customer_gender_other" json:"customer_gender_other"`
	CustomerBirthdate                     time.Time                    `bson:"customer_birthdate" json:"customer_birthdate"`
	CustomerEmail                         string                       `bson:"customer_email" json:"customer_email"`
	CustomerPhone                         string                       `bson:"customer_phone" json:"customer_phone,omitempty"`
	CustomerPhoneType                     int8                         `bson:"customer_phone_type" json:"customer_phone_type"`
	CustomerPhoneExtension                string                       `bson:"customer_phone_extension" json:"customer_phone_extension"`
	CustomerOtherPhone                    string                       `bson:"customer_other_phone" json:"customer_other_phone"`
	CustomerOtherPhoneExtension           string                       `bson:"customer_other_phone_extension" json:"customer_other_phone_extension"`
	CustomerOtherPhoneType                int8                         `bson:"customer_other_phone_type" json:"customer_other_phone_type"`
	CustomerFullAddressWithoutPostalCode  string                       `bson:"customer_full_address_without_postal_code" json:"customer_full_address_without_postal_code"`
	CustomerFullAddressURL                string                       `bson:"customer_full_address_url" json:"customer_full_address_url"`
	CustomerTags                          []*OrderTag                  `bson:"customer_tags" json:"customer_tags,omitempty"`
	AssociateID                           primitive.ObjectID           `bson:"associate_id" json:"associate_id"`
	AssociateFirstName                    string                       `bson:"associate_first_name" json:"associate_first_name,omitempty"`
	AssociateLastName                     string                       `bson:"associate_last_name" json:"associate_last_name,omitempty"`
	AssociateName                         string                       `bson:"associate_name" json:"associate_name,omitempty"`
	AssociateLexicalName                  string                       `bson:"associate_lexical_name" json:"associate_lexical_name,omitempty"`
	AssociateGender                       int8                         `bson:"associate_gender" json:"associate_gender"`
	AssociateGenderOther                  string                       `bson:"associate_gender_other" json:"associate_gender_other"`
	AssociateBirthdate                    time.Time                    `bson:"associate_birthdate" json:"associate_birthdate"`
	AssociateEmail                        string                       `bson:"associate_email" json:"associate_email,omitempty"`
	AssociatePhone                        string                       `bson:"associate_phone" json:"associate_phone,omitempty"`
	AssociatePhoneType                    int8                         `bson:"associate_phone_type" json:"associate_phone_type"`
	AssociatePhoneExtension               string                       `bson:"associate_phone_extension" json:"associate_phone_extension"`
	AssociateOtherPhone                   string                       `bson:"associate_other_phone" json:"associate_other_phone"`
	AssociateOtherPhoneExtension          string                       `bson:"associate_other_phone_extension" json:"associate_other_phone_extension"`
	AssociateOtherPhoneType               int8                         `bson:"associate_other_phone_type" json:"associate_other_phone_type"`
	AssociateFullAddressWithoutPostalCode string                       `bson:"associate_full_address_without_postal_code" json:"associate_full_address_without_postal_code"`
	AssociateFullAddressURL               string                       `bson:"associate_full_address_url" json:"associate_full_address_url"`
	AssociateTags                         []*OrderTag                  `bson:"associate_tags" json:"associate_tags,omitempty"`
	AssociateSkillSets                    []*OrderSkillSet             `bson:"associate_skill_sets" json:"associate_skill_sets,omitempty"`
	AssociateInsuranceRequirements        []*OrderInsuranceRequirement `bson:"associate_insurance_requirements" json:"associate_insurance_requirements,omitempty"`
	AssociateVehicleTypes                 []*OrderVehicleType          `bson:"associate_vehicle_types" json:"associate_vehicle_types,omitempty"`
	AssociateTaxID                        string                       `bson:"associate_tax_id" json:"associate_tax_id"`
	TenantID                              primitive.ObjectID           `bson:"tenant_id" json:"tenant_id,omitempty"`
	TenantIDWithWJID                      string                       `bson:"tenant_id_with_wjid" json:"tenant_id_with_wjid"` // TenantIDWithWJID is a combination of `tenancy_id` and `wjid` values written in the following structure `%v_%v`.
	Description                           string                       `bson:"description" json:"description"`
	AssignmentDate                        time.Time                    `bson:"assignment_date" json:"assignment_date"`
	IsOngoing                             bool                         `bson:"is_ongoing" json:"is_ongoing"`
	IsHomeSupportService                  bool                         `bson:"is_home_support_service" json:"is_home_support_service"`
	StartDate                             time.Time                    `bson:"start_date" json:"start_date"`
	CompletionDate                        time.Time                    `bson:"completion_date" json:"completion_date"`
	Hours                                 float64                      `bson:"hours" json:"hours"`
	Type                                  int8                         `bson:"type" json:"type"`
	IndexedText                           string                       `bson:"indexed_text" json:"indexed_text"`
	ClosingReason                         int8                         `bson:"closing_reason" json:"closing_reason"`
	ClosingReasonOther                    string                       `bson:"closing_reason_other" json:"closing_reason_other"`
	Status                                int8                         `bson:"status" json:"status"`
	Currency                              string                       `bson:"currency" json:"currency"`
	WasJobSatisfactory                    bool                         `bson:"was_job_satisfactory" json:"was_job_satisfactory"`
	WasJobFinishedOnTimeAndOnBudget       bool                         `bson:"was_job_finished_on_time_and_on_budget" json:"was_job_finished_on_time_and_on_budget"`
	WasAssociatePunctual                  bool                         `bson:"was_associate_punctual" json:"was_associate_punctual"`
	WasAssociateProfessional              bool                         `bson:"was_associate_professional" json:"was_associate_professional"`
	WouldCustomerReferOurOrganization     bool                         `bson:"would_customer_refer_our_organization" json:"would_customer_refer_our_organization"`
	Score                                 float64                      `bson:"score" json:"score"`
	InvoiceDate                           time.Time                    `bson:"invoice_date" json:"invoice_date"`
	InvoiceQuoteAmount                    float64                      `bson:"invoice_quote_amount" json:"invoice_quote_amount"`
	InvoiceLabourAmount                   float64                      `bson:"invoice_labour_amount" json:"invoice_labour_amount"`
	InvoiceMaterialAmount                 float64                      `bson:"invoice_material_amount" json:"invoice_material_amount"`
	InvoiceTaxAmount                      float64                      `bson:"invoice_tax_amount" json:"invoice_tax_amount"`
	InvoiceTotalAmount                    float64                      `bson:"invoice_total_amount" json:"invoice_total_amount"`
	InvoiceServiceFeeAmount               float64                      `bson:"invoice_service_fee_amount" json:"invoice_service_fee_amount"`
	InvoiceServiceFeePaymentDate          time.Time                    `bson:"invoice_service_fee_payment_date" json:"invoice_service_fee_payment_date"`
	CreatedAt                             time.Time                    `bson:"created_at" json:"created_at"`
	CreatedByUserID                       primitive.ObjectID           `bson:"created_by_user_id" json:"created_by_user_id,omitempty"`
	CreatedByUserName                     string                       `bson:"created_by_user_name" json:"created_by_user_name"`
	CreatedFromIPAddress                  string                       `bson:"created_from_ip_address" json:"created_from_ip_address"`
	ModifiedAt                            time.Time                    `bson:"modified_at" json:"modified_at"`
	ModifiedByUserID                      primitive.ObjectID           `bson:"modified_by_user_id" json:"modified_by_user_id,omitempty"`
	ModifiedByUserName                    string                       `bson:"modified_by_user_name" json:"modified_by_user_name"`
	ModifiedFromIPAddress                 string                       `bson:"modified_from_ip_address" json:"modified_from_ip_address"`
	InvoiceServiceFeeID                   primitive.ObjectID           `bson:"invoice_service_fee_id" json:"invoice_service_fee_id"`
	InvoiceServiceFeeName                 string                       `bson:"invoice_service_fee_name" json:"invoice_service_fee_name"`
	InvoiceServiceFeeDescription          string                       `bson:"invoice_service_fee_description" json:"invoice_service_fee_description"`
	InvoiceServiceFeePercentage           float64                      `bson:"invoice_service_fee_percentage" json:"invoice_service_fee_percentage"`
	LatestPendingTaskID                   primitive.ObjectID           `bson:"latest_pending_task_id" json:"latest_pending_task_id"`
	LatestPendingTaskTitle                string                       `bson:"latest_pending_task_title" json:"latest_pending_task_title"`
	LatestPendingTaskDescription          string                       `bson:"latest_pending_task_description" json:"latest_pending_task_description"`
	LatestPendingTaskDueDate              time.Time                    `bson:"latest_pending_task_due_date" json:"latest_pending_task_due_date"`
	LatestPendingTaskType                 int8                         `bson:"latest_pending_task_type" json:"latest_pending_task_type"`
	OngoingOrderID                        primitive.ObjectID           `bson:"ongoing_work_order_id" json:"ongoing_work_order_id"`
	WasSurveyConducted                    bool                         `bson:"was_survey_conducted" json:"was_survey_conducted"`
	WasThereFinancialsInputted            bool                         `bson:"was_there_financials_inputted" json:"was_there_financials_inputted"`
	InvoiceActualServiceFeeAmountPaid     float64                      `bson:"invoice_actual_service_fee_amount_paid" json:"invoice_actual_service_fee_amount_paid"`
	InvoiceBalanceOwingAmount             float64                      `bson:"invoice_balance_owing_amount" json:"invoice_balance_owing_amount"`
	InvoiceQuotedLabourAmount             float64                      `bson:"invoice_quoted_labour_amount" json:"invoice_quoted_labour_amount"`
	InvoiceQuotedMaterialAmount           float64                      `bson:"invoice_quoted_material_amount" json:"invoice_quoted_material_amount"`
	InvoiceTotalQuoteAmount               float64                      `bson:"invoice_total_quote_amount" json:"invoice_total_quote_amount"`
	Visits                                int8                         `bson:"visits" json:"visits"`
	InvoiceIDs                            string                       `bson:"invoice_ids" json:"invoice_ids"`
	NoSurveyConductedReason               int8                         `bson:"no_survey_conducted_reason" json:"no_survey_conducted_reason"`
	NoSurveyConductedReasonOther          string                       `bson:"no_survey_conducted_reason_other" json:"no_survey_conducted_reason_other"`
	ClonedFromOrderID                     primitive.ObjectID           `bson:"cloned_from_order_id" json:"cloned_from_order_id"`
	InvoiceDepositAmount                  float64                      `bson:"invoice_deposit_amount_id" json:"invoice_deposit_amount"`
	InvoiceOtherCostsAmount               float64                      `bson:"invoice_other_costs_amount" json:"invoice_other_costs_amount"`
	InvoiceQuotedOtherCostsAmount         float64                      `bson:"invoice_quoted_other_costs_amount" json:"invoice_quoted_other_costs_amount"`
	InvoicePaidTo                         int8                         `bson:"invoice_paid_to" json:"invoice_paid_to"`
	InvoiceAmountDue                      float64                      `bson:"invoice_amount_due" json:"invoice_amount_due"`
	InvoiceSubTotalAmount                 float64                      `bson:"invoice_sub_total_amount" json:"invoice_sub_total_amount"`
	ClosingReasonComment                  string                       `bson:"closing_reason_comment" json:"closing_reason_comment"`
	Tags                                  []*OrderTag                  `bson:"tags" json:"tags,omitempty"`
	SkillSets                             []*OrderSkillSet             `bson:"skill_sets" json:"skill_sets,omitempty"`
	Comments                              []*OrderComment              `bson:"comments" json:"comments,omitempty"`
	Deposits                              []*OrderDeposit              `bson:"deposits" json:"deposits,omitempty"`
	Invoice                               *OrderInvoice                `bson:"invoice" json:"invoice,omitempty"`
	PastInvoices                          []*OrderInvoice              `bson:"past_invoices" json:"past_invoices,omitempty"`
}

type OrderComment struct {
	ID                    primitive.ObjectID `bson:"_id" json:"id"`
	OrderID               primitive.ObjectID `bson:"order_id" json:"order_id"`
	OrderWJID             uint64             `bson:"order_wjid" json:"order_wjid"`                               // Workery Job ID
	OrderTenantIDWithWJID string             `bson:"order_tenant_id_with_wjid" json:"order_tenant_id_with_wjid"` // OrderTenantIDWithWJID is a combination of `tenancy_id` and `wjid` values written in the following structure `%v_%v`.
	TenantID              primitive.ObjectID `bson:"tenant_id" json:"tenant_id"`
	CreatedAt             time.Time          `bson:"created_at,omitempty" json:"created_at,omitempty"`
	CreatedByUserID       primitive.ObjectID `bson:"created_by_user_id" json:"created_by_user_id"`
	CreatedByUserName     string             `bson:"created_by_user_name" json:"created_by_user_name"`
	CreatedFromIPAddress  string             `bson:"created_from_ip_address" json:"created_from_ip_address"`
	ModifiedAt            time.Time          `bson:"modified_at,omitempty" json:"modified_at,omitempty"`
	ModifiedByUserID      primitive.ObjectID `bson:"modified_by_user_id" json:"modified_by_user_id"`
	ModifiedByUserName    string             `bson:"modified_by_user_name" json:"modified_by_user_name"`
	ModifiedFromIPAddress string             `bson:"modified_from_ip_address" json:"modified_from_ip_address"`
	Content               string             `bson:"content" json:"content"`
	Status                int8               `bson:"status" json:"status"`
	PublicID              uint64             `bson:"public_id" json:"public_id"` // Workery Job ID
}

type OrderInvoice struct {
	ID                       primitive.ObjectID `bson:"_id" json:"id"`
	OrderID                  primitive.ObjectID `bson:"order_id" json:"order_id"`
	OrderWJID                uint64             `bson:"order_wjid" json:"order_wjid"`                               // Workery Job ID
	OrderTenantIDWithWJID    string             `bson:"order_tenant_id_with_wjid" json:"order_tenant_id_with_wjid"` // OrderTenantIDWithWJID is a combination of `tenancy_id` and `wjid` values written in the following structure `%v_%v`.
	TenantID                 primitive.ObjectID `bson:"tenant_id" json:"tenant_id"`
	CreatedAt                time.Time          `bson:"created_at,omitempty" json:"created_at,omitempty"`
	CreatedByUserID          primitive.ObjectID `bson:"created_by_user_id" json:"created_by_user_id"`
	CreatedByUserName        string             `bson:"created_by_user_name" json:"created_by_user_name"`
	CreatedFromIPAddress     string             `bson:"created_from_ip_address" json:"created_from_ip_address"`
	ModifiedAt               time.Time          `bson:"modified_at,omitempty" json:"modified_at,omitempty"`
	ModifiedByUserID         primitive.ObjectID `bson:"modified_by_user_id" json:"modified_by_user_id"`
	ModifiedByUserName       string             `bson:"modified_by_user_name" json:"modified_by_user_name"`
	ModifiedFromIPAddress    string             `bson:"modified_from_ip_address" json:"modified_from_ip_address"`
	InvoiceID                string             `bson:"invoice_id" json:"invoice_id"`
	InvoiceDate              time.Time          `bson:"invoice_date" json:"invoice_date"`
	AssociateName            string             `bson:"associate_name" json:"associate_name"`
	AssociatePhone           string             `bson:"associate_phone" json:"associate_phone"`
	ClientName               string             `bson:"client_name" json:"client_name"`
	ClientPhone              string             `bson:"client_phone" json:"client_phone"`
	ClientEmail              string             `bson:"client_email" json:"client_email"`
	Line01Qty                int64              `bson:"line_01_qty" json:"line_01_qty"`
	Line01Desc               string             `bson:"line_01_desc" json:"line_01_desc"`
	Line01Price              float64            `bson:"line_01_price" json:"line_01_price"`
	Line01Amount             float64            `bson:"line_01_amount" json:"line_01_amount"`
	Line02Qty                int64              `bson:"line_02_qty" json:"line_02_qty"` // Make `int8`
	Line02Desc               string             `bson:"line_02_desc" json:"line_02_desc"`
	Line02Price              float64            `bson:"line_02_price" json:"line_02_price"`
	Line02Amount             float64            `bson:"line_02_amount" json:"line_02_amount"`
	Line03Qty                int64              `bson:"line_03_qty" json:"line_03_qty"` // Make `int8`
	Line03Desc               string             `bson:"line_03_desc" json:"line_03_desc"`
	Line03Price              float64            `bson:"line_03_price" json:"line_03_price"`
	Line03Amount             float64            `bson:"line_03_amount" json:"line_03_amount"`
	Line04Qty                int64              `bson:"line_04_qty" json:"line_04_qty"` // Make `int8`
	Line04Desc               string             `bson:"line_04_desc" json:"line_04_desc"`
	Line04Price              float64            `bson:"line_04_price" json:"line_04_price"`
	Line04Amount             float64            `bson:"line_04_amount" json:"line_04_amount"`
	Line05Qty                int64              `bson:"line_05_qty" json:"line_05_qty"` // Make `int8`
	Line05Desc               string             `bson:"line_05_desc" json:"line_05_desc"`
	Line05Price              float64            `bson:"line_05_price" json:"line_05_price"`
	Line05Amount             float64            `bson:"line_05_amount" json:"line_05_amount"`
	Line06Qty                int64              `bson:"line_06_qty" json:"line_06_qty"` // Make `int8`
	Line06Desc               string             `bson:"line_06_desc" json:"line_06_desc"`
	Line06Price              float64            `bson:"line_06_price" json:"line_06_price"`
	Line06Amount             float64            `bson:"line_06_amount" json:"line_06_amount"`
	Line07Qty                int64              `bson:"line_07_qty" json:"line_07_qty"` // Make `int8`
	Line07Desc               string             `bson:"line_07_desc" json:"line_07_desc"`
	Line07Price              float64            `bson:"line_07_price" json:"line_07_price"`
	Line07Amount             float64            `bson:"line_07_amount" json:"line_07_amount"`
	Line08Qty                int64              `bson:"line_08_qty" json:"line_08_qty"` // Make `int8`
	Line08Desc               string             `bson:"line_08_desc" json:"line_08_desc"`
	Line08Price              float64            `bson:"line_08_price" json:"line_08_price"`
	Line08Amount             float64            `bson:"line_08_amount" json:"line_08_amount"`
	Line09Qty                int64              `bson:"line_09_qty" json:"line_09_qty"` // Make `int8`
	Line09Desc               string             `bson:"line_09_desc" json:"line_09_desc"`
	Line09Price              float64            `bson:"line_09_price" json:"line_09_price"`
	Line09Amount             float64            `bson:"line_09_amount" json:"line_09_amount"`
	Line10Qty                int64              `bson:"line_10_qty" json:"line_10_qty"` // Make `int8`
	Line10Desc               string             `bson:"line_10_desc" json:"line_10_desc"`
	Line10Price              float64            `bson:"line_10_price" json:"line_10_price"`
	Line10Amount             float64            `bson:"line_10_amount" json:"line_10_amount"`
	Line11Qty                int64              `bson:"line_11_qty" json:"line_11_qty"` // Make `int8`
	Line11Desc               string             `bson:"line_11_desc" json:"line_11_desc"`
	Line11Price              float64            `bson:"line_11_price" json:"line_11_price"`
	Line11Amount             float64            `bson:"line_11_amount" json:"line_11_amount"`
	Line12Qty                int64              `bson:"line_12_qty" json:"line_12_qty"` // Make `int8`
	Line12Desc               string             `bson:"line_12_desc" json:"line_12_desc"`
	Line12Price              float64            `bson:"line_12_price" json:"line_12_price"`
	Line12Amount             float64            `bson:"line_12_amount" json:"line_12_amount"`
	Line13Qty                int64              `bson:"line_13_qty" json:"line_13_qty"` // Make `int8`
	Line13Desc               string             `bson:"line_13_desc" json:"line_13_desc"`
	Line13Price              float64            `bson:"line_13_price" json:"line_13_price"`
	Line13Amount             float64            `bson:"line_13_amount" json:"line_13_amount"`
	Line14Qty                int64              `bson:"line_14_qty" json:"line_14_qty"` // Make `int8`
	Line14Desc               string             `bson:"line_14_desc" json:"line_14_desc"`
	Line14Price              float64            `bson:"line_14_price" json:"line_14_price"`
	Line14Amount             float64            `bson:"line_14_amount" json:"line_14_amount"`
	Line15Qty                int64              `bson:"line_15_qty" json:"line_15_qty"` // Make `int8`
	Line15Desc               string             `bson:"line_15_desc" json:"line_15_desc"`
	Line15Price              float64            `bson:"line_15_price" json:"line_15_price"`
	Line15Amount             float64            `bson:"line_15_amount" json:"line_15_amount"`
	InvoiceQuoteDays         int8               `bson:"invoice_quote_days" json:"invoice_quote_days"`
	InvoiceAssociateTax      string             `bson:"invoice_associate_tax" json:"invoice_associate_tax"`
	InvoiceQuoteDate         time.Time          `bson:"invoice_quote_date" json:"invoice_quote_date"`
	InvoiceCustomersApproval string             `bson:"invoice_customers_approval" json:"invoice_customers_approval"`
	Line01Notes              string             `bson:"line_01_notes" json:"line_01_notes"`
	Line02Notes              string             `bson:"line_02_notes" json:"line_02_notes"`
	TotalLabour              float64            `bson:"total_labour" json:"total_labour"`
	TotalMaterials           float64            `bson:"total_materials" json:"total_materials"`
	OtherCosts               float64            `bson:"other_costs" json:"other_costs"`
	Tax                      float64            `bson:"tax" json:"tax"`
	Total                    float64            `bson:"total" json:"total"`
	PaymentAmount            float64            `bson:"payment_amount" json:"payment_amount"`
	PaymentDate              time.Time          `bson:"payment_date" json:"payment_date"`
	IsCash                   bool               `bson:"is_cash" json:"is_cash"`
	IsCheque                 bool               `bson:"is_cheque" json:"is_cheque"`
	IsDebit                  bool               `bson:"is_debit" json:"is_debit"`
	IsCredit                 bool               `bson:"is_credit" json:"is_credit"`
	IsOther                  bool               `bson:"is_other" json:"is_other"`
	ClientSignature          string             `bson:"client_signature" json:"client_signature"`
	AssociateSignDate        time.Time          `bson:"associate_sign_date" json:"associate_sign_date"`
	AssociateSignature       string             `bson:"associate_signature" json:"associate_signature"`
	WorkOrderID              uint64             `bson:"work_order_id" json:"work_order_id"`
	ClientAddress            string             `bson:"client_address" json:"client_address"`
	RevisionVersion          int8               `bson:"revision_version" json:"revision_version"`
	Deposit                  float64            `bson:"deposit" json:"deposit"`
	AmountDue                float64            `bson:"amount_due" json:"amount_due"`
	SubTotal                 float64            `bson:"sub_total" json:"sub_total"`
	FileObjectKey            string             `bson:"file_object_key" json:"file_object_key"`
	FileTitle                string             `bson:"file_title" json:"file_title"`
	FileObjectURL            string             `bson:"file_object_url" json:"file_object_url,omitempty"`       // (Optional, added by endpoint)
	FileObjectExpiry         time.Time          `bson:"file_object_expiry" json:"file_object_expiry,omitempty"` // (Optional, added by endpoint)
	Status                   int8               `bson:"status" json:"status"`
	PublicID                 uint64             `bson:"public_id" json:"public_id"`
}

type OrderDeposit struct {
	ID                    primitive.ObjectID `bson:"_id" json:"id"`
	OrderID               primitive.ObjectID `bson:"order_id" json:"order_id"`
	OrderWJID             uint64             `bson:"order_wjid" json:"order_wjid"`                               // Workery Job ID
	OrderTenantIDWithWJID string             `bson:"order_tenant_id_with_wjid" json:"order_tenant_id_with_wjid"` // OrderTenantIDWithWJID is a combination of `tenancy_id` and `wjid` values written in the following structure `%v_%v`.
	TenantID              primitive.ObjectID `bson:"tenant_id" json:"tenant_id"`
	PaidAt                time.Time          `bson:"paid_at,omitempty" json:"paid_at"`
	DepositMethod         int8               `bson:"deposit_method" json:"deposit_method"`
	PaidTo                int8               `bson:"paid_to,omitempty" json:"paid_to"`
	Currency              string             `bson:"currency" json:"currency"`
	Amount                float64            `bson:"amount" json:"amount"`
	PaidFor               int8               `bson:"paid_for" json:"paid_for"`
	CreatedAt             time.Time          `bson:"created_at" json:"created_at"`
	CreatedByUserID       primitive.ObjectID `bson:"created_by_user_id" json:"created_by_user_id,omitempty"`
	CreatedByUserName     string             `bson:"created_by_user_name" json:"created_by_user_name"`
	CreatedFromIPAddress  string             `bson:"created_from_ip_address" json:"created_from_ip_address"`
	ModifiedAt            time.Time          `bson:"modified_at" json:"modified_at"`
	ModifiedByUserID      primitive.ObjectID `bson:"modified_by_user_id" json:"modified_by_user_id,omitempty"`
	ModifiedByUserName    string             `bson:"modified_by_user_name" json:"modified_by_user_name"`
	ModifiedFromIPAddress string             `bson:"modified_from_ip_address" json:"modified_from_ip_address"`
	Status                int8               `bson:"status" json:"status"`
	PublicID              uint64             `bson:"public_id" json:"public_id"`
}

type OrderTag struct {
	ID          primitive.ObjectID `bson:"_id" json:"id"`
	Text        string             `bson:"text" json:"text"`
	Description string             `bson:"description" json:"description"`
	Status      int8               `bson:"status" json:"status"`
}

type OrderSkillSet struct {
	ID          primitive.ObjectID `bson:"id" json:"id"`
	Category    string             `bson:"category" json:"category,omitempty"`         // Referenced value from 'tags'.
	SubCategory string             `bson:"sub_category" json:"sub_category,omitempty"` // Referenced value from 'tags'.
	Description string             `bson:"description" json:"description,omitempty"`   // Referenced value from 'tags'.
	Status      int8               `bson:"status" json:"status"`
}

type OrderInsuranceRequirement struct {
	ID          primitive.ObjectID `bson:"_id" json:"id"`
	Name        string             `bson:"name" json:"name"`
	Description string             `bson:"description" json:"description"`
	Status      int8               `bson:"status" json:"status"`
}

type OrderVehicleType struct {
	ID          primitive.ObjectID `bson:"_id" json:"id"`
	Name        string             `bson:"name" json:"name"`
	Description string             `bson:"description" json:"description"`
	Status      int8               `bson:"status" json:"status"`
}

type OrderListFilter struct {
	// Pagination related.
	Cursor    primitive.ObjectID
	PageSize  int64
	SortField string
	SortOrder int8 // 1=ascending | -1=descending

	// Filter related.
	TenantID         primitive.ObjectID
	CustomerID       primitive.ObjectID
	AssociateID      primitive.ObjectID
	Status           int8
	Type             int8
	ExcludeArchived  bool
	SearchText       string
	ModifiedByUserID primitive.ObjectID
}

type OrderListResult struct {
	Results     []*Order           `json:"results"`
	NextCursor  primitive.ObjectID `json:"next_cursor"`
	HasNextPage bool               `json:"has_next_page"`
}

type OrderAsSelectOption struct {
	Value primitive.ObjectID `bson:"_id" json:"value"` // Extract from the database `_id` field and output through API as `value`.
	Label string             `bson:"name" json:"label"`
}

// OrderStorer Interface for user.
type OrderStorer interface {
	Create(ctx context.Context, m *Order) error
	GetByID(ctx context.Context, id primitive.ObjectID) (*Order, error)
	GetByWJID(ctx context.Context, wjid uint64) (*Order, error)
	GetLatestOrderByTenantID(ctx context.Context, tenantID primitive.ObjectID) (*Order, error)
	// GetByPublicID(ctx context.Context, oldID uint64) (*Order, error)
	// GetByEmail(ctx context.Context, email string) (*Order, error)
	// GetByVerificationCode(ctx context.Context, verificationCode string) (*Order, error)
	// CheckIfExistsByEmail(ctx context.Context, email string) (bool, error)
	UpdateByID(ctx context.Context, m *Order) error
	ListByFilter(ctx context.Context, f *OrderPaginationListFilter) (*OrderPaginationListResult, error)
	LiteListByFilter(ctx context.Context, f *OrderPaginationListFilter) (*OrderPaginationLiteListResult, error)
	ListByCustomerID(ctx context.Context, customerID primitive.ObjectID) (*OrderPaginationListResult, error)
	ListByAssociateID(ctx context.Context, associateID primitive.ObjectID) (*OrderPaginationListResult, error)
	ListByServiceFeeID(ctx context.Context, serviceFeeID primitive.ObjectID) (*OrderPaginationListResult, error)
	// ListAsSelectOptionByFilter(ctx context.Context, f *OrderListFilter) ([]*OrderAsSelectOption, error)
	// DeleteByID(ctx context.Context, id primitive.ObjectID) error
	CountByFilter(ctx context.Context, f *OrderListFilter) (int64, error)
	CountByAssociateID(ctx context.Context, associateID primitive.ObjectID) (int64, error)
	CountByTenantID(ctx context.Context, tenantID primitive.ObjectID) (int64, error)
	PermanentlyDeleteAllByCustomerID(ctx context.Context, customerID primitive.ObjectID) error
	PermanentlyDeleteAllByAssociateID(ctx context.Context, associateID primitive.ObjectID) error
}

type OrderStorerImpl struct {
	Logger     *slog.Logger
	DbClient   *mongo.Client
	Collection *mongo.Collection
}

func NewDatastore(appCfg *c.Conf, loggerp *slog.Logger, client *mongo.Client) OrderStorer {
	// ctx := context.Background()
	uc := client.Database(appCfg.DB.Name).Collection("orders")

	// // For debugging purposes only.
	// if _, err := uc.Indexes().DropAll(context.TODO()); err != nil {
	// 	loggerp.Error("failed deleting all indexes",
	// 		slog.Any("err", err))
	//
	// 	// It is important that we crash the app on startup to meet the
	// 	// requirements of `google/wire` framework.
	// 	log.Fatal(err)
	// }

	_, err := uc.Indexes().CreateMany(context.TODO(), []mongo.IndexModel{
		{Keys: bson.D{{Key: "tenant_id", Value: 1}}},
		{Keys: bson.D{{Key: "wjid", Value: -1}}},
		{Keys: bson.D{{Key: "tenant_id_with_wjid", Value: 1}}},
		{Keys: bson.D{{Key: "customer_id", Value: 1}}},
		{Keys: bson.D{{Key: "associate_id", Value: 1}}},
		{Keys: bson.D{{Key: "start_date", Value: 1}}},
		{Keys: bson.D{{Key: "completion_date", Value: 1}}},
		{Keys: bson.D{{Key: "assignment_date", Value: -1}}},
		{Keys: bson.D{{Key: "status", Value: 1}}},
		{Keys: bson.D{{Key: "type", Value: 1}}},
		{Keys: bson.D{{Key: "customer_lexical_name", Value: 1}}},
		{Keys: bson.D{{Key: "associate_lexical_name", Value: 1}}},
		{Keys: bson.D{
			{"customer_name", "text"},
			{"customer_lexical_name", "text"},
			{"customer_email", "text"},
			{"customer_phone", "text"},
			{"customer_other_phone", "text"},
			{"customer_full_address_without_postal_code", "text"},
			{"customer_tags", "text"},
			{"associate_name", "text"},
			{"associate_lexical_name", "text"},
			{"associate_email", "text"},
			{"associate_phone", "text"},
			{"associate_other_phone", "text"},
			{"associate_full_address_without_postal_code", "text"},
			{"associate_tags", "text"},
			{"associate_skill_sets", "text"},
			{"associate_insurance_requirements", "text"},
			{"associate_vehicle_types", "text"},
			{"tenant_id_with_wjid", "text"},
			{"description", "text"},
			{"closing_reason_other", "text"},
			{"invoice_service_fee_name", "text"},
			{"invoice_service_fee_description", "text"},
			{"latest_pending_task_description", "text"},
			{"no_survey_conducted_reason_other", "text"},
			{"tags", "text"},
			{"skill_sets", "text"},
			{"comments", "text"},
		}},
	})
	if err != nil {
		// It is important that we crash the app on startup to meet the
		// requirements of `google/wire` framework.
		log.Fatal(err)
	}

	s := &OrderStorerImpl{
		Logger:     loggerp,
		DbClient:   client,
		Collection: uc,
	}
	return s
}
