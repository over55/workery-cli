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
	o_ds "github.com/over55/workery-cli/app/order/datastore"
	tenant_ds "github.com/over55/workery-cli/app/tenant/datastore"
	user_ds "github.com/over55/workery-cli/app/user/datastore"
	"github.com/over55/workery-cli/config"
)

func init() {
	rootCmd.AddCommand(importOrderInvoiceCmd)
}

var importOrderInvoiceCmd = &cobra.Command{
	Use:   "import_order_invoice",
	Short: "Import the order invoices from old database",
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

		tenant, err := tenantStorer.GetBySchemaName(context.Background(), cfg.PostgresDB.DatabaseLondonSchemaName)
		if err != nil {
			log.Fatal(err)
		}

		RunImportOrderInvoice(cfg, ppc, lpc, tenantStorer, userStorer, oStorer, tenant)
	},
}

type OldWorkOrderInvoice struct {
	OrderID                  uint64      `json:"order_id"`
	IsArchived               bool        `json:"is_archived"`
	InvoiceID                string      `json:"invoice_id"`
	InvoiceDate              time.Time   `json:"invoice_date"`
	AssociateName            string      `json:"associate_name"`
	AssociateTelephone       string      `json:"associate_telephone"`
	ClientName               string      `json:"client_name"`
	ClientTelephone          string      `json:"client_telephone"`
	ClientEmail              null.String `json:"client_email"`
	Line01Qty                int8        `json:"line_01_qty"`
	Line01Desc               string      `json:"line_01_desc"`
	Line01PriceCurrency      string      `json:"line_01_price_currency"`
	Line01Price              float64     `json:"line_01_price"`
	Line01AmountCurrency     string      `json:"line_01_amount_currency"`
	Line01Amount             float64     `json:"line_01_amount"`
	Line02Qty                null.Int    `json:"line_02_qty"` // Make `int8`
	Line02Desc               null.String `json:"line_02_desc"`
	Line02PriceCurrency      null.String `json:"line_02_price_currency"`
	Line02Price              null.Float  `json:"line_02_price"`
	Line02AmountCurrency     null.String `json:"line_02_amount_currency"`
	Line02Amount             null.Float  `json:"line_02_amount"`
	Line03Qty                null.Int    `json:"line_03_qty"` // Make `int8`
	Line03Desc               null.String `json:"line_03_desc"`
	Line03PriceCurrency      null.String `json:"line_03_price_currency"`
	Line03Price              null.Float  `json:"line_03_price"`
	Line03AmountCurrency     null.String `json:"line_03_amount_currency"`
	Line03Amount             null.Float  `json:"line_03_amount"`
	Line04Qty                null.Int    `json:"line_04_qty"` // Make `int8`
	Line04Desc               null.String `json:"line_04_desc"`
	Line04PriceCurrency      null.String `json:"line_04_price_currency"`
	Line04Price              null.Float  `json:"line_04_price"`
	Line04AmountCurrency     null.String `json:"line_04_amount_currency"`
	Line04Amount             null.Float  `json:"line_04_amount"`
	Line05Qty                null.Int    `json:"line_05_qty"` // Make `int8`
	Line05Desc               null.String `json:"line_05_desc"`
	Line05PriceCurrency      null.String `json:"line_05_price_currency"`
	Line05Price              null.Float  `json:"line_05_price"`
	Line05AmountCurrency     null.String `json:"line_05_amount_currency"`
	Line05Amount             null.Float  `json:"line_05_amount"`
	Line06Qty                null.Int    `json:"line_06_qty"` // Make `int8`
	Line06Desc               null.String `json:"line_06_desc"`
	Line06PriceCurrency      null.String `json:"line_06_price_currency"`
	Line06Price              null.Float  `json:"line_06_price"`
	Line06AmountCurrency     null.String `json:"line_06_amount_currency"`
	Line06Amount             null.Float  `json:"line_06_amount"`
	Line07Qty                null.Int    `json:"line_07_qty"` // Make `int8`
	Line07Desc               null.String `json:"line_07_desc"`
	Line07PriceCurrency      null.String `json:"line_07_price_currency"`
	Line07Price              null.Float  `json:"line_07_price"`
	Line07AmountCurrency     null.String `json:"line_07_amount_currency"`
	Line07Amount             null.Float  `json:"line_07_amount"`
	Line08Qty                null.Int    `json:"line_08_qty"` // Make `int8`
	Line08Desc               null.String `json:"line_08_desc"`
	Line08PriceCurrency      null.String `json:"line_08_price_currency"`
	Line08Price              null.Float  `json:"line_08_price"`
	Line08AmountCurrency     null.String `json:"line_08_amount_currency"`
	Line08Amount             null.Float  `json:"line_08_amount"`
	Line09Qty                null.Int    `json:"line_09_qty"` // Make `int8`
	Line09Desc               null.String `json:"line_09_desc"`
	Line09PriceCurrency      null.String `json:"line_09_price_currency"`
	Line09Price              null.Float  `json:"line_09_price"`
	Line09AmountCurrency     null.String `json:"line_09_amount_currency"`
	Line09Amount             null.Float  `json:"line_09_amount"`
	Line10Qty                null.Int    `json:"line_10_qty"` // Make `int8`
	Line10Desc               null.String `json:"line_10_desc"`
	Line10PriceCurrency      null.String `json:"line_10_price_currency"`
	Line10Price              null.Float  `json:"line_10_price"`
	Line10AmountCurrency     null.String `json:"line_10_amount_currency"`
	Line10Amount             null.Float  `json:"line_10_amount"`
	Line11Qty                null.Int    `json:"line_11_qty"` // Make `int8`
	Line11Desc               null.String `json:"line_11_desc"`
	Line11PriceCurrency      null.String `json:"line_11_price_currency"`
	Line11Price              null.Float  `json:"line_11_price"`
	Line11AmountCurrency     null.String `json:"line_11_amount_currency"`
	Line11Amount             null.Float  `json:"line_11_amount"`
	Line12Qty                null.Int    `json:"line_12_qty"` // Make `int8`
	Line12Desc               null.String `json:"line_12_desc"`
	Line12PriceCurrency      null.String `json:"line_12_price_currency"`
	Line12Price              null.Float  `json:"line_12_price"`
	Line12AmountCurrency     null.String `json:"line_12_amount_currency"`
	Line12Amount             null.Float  `json:"line_12_amount"`
	Line13Qty                null.Int    `json:"line_13_qty"` // Make `int8`
	Line13Desc               null.String `json:"line_13_desc"`
	Line13PriceCurrency      null.String `json:"line_13_price_currency"`
	Line13Price              null.Float  `json:"line_13_price"`
	Line13AmountCurrency     null.String `json:"line_13_amount_currency"`
	Line13Amount             null.Float  `json:"line_13_amount"`
	Line14Qty                null.Int    `json:"line_14_qty"` // Make `int8`
	Line14Desc               null.String `json:"line_14_desc"`
	Line14PriceCurrency      null.String `json:"line_14_price_currency"`
	Line14Price              null.Float  `json:"line_14_price"`
	Line14AmountCurrency     null.String `json:"line_14_amount_currency"`
	Line14Amount             null.Float  `json:"line_14_amount"`
	Line15Qty                null.Int    `json:"line_15_qty"` // Make `int8`
	Line15Desc               null.String `json:"line_15_desc"`
	Line15PriceCurrency      null.String `json:"line_15_price_currency"`
	Line15Price              null.Float  `json:"line_15_price"`
	Line15AmountCurrency     null.String `json:"line_15_amount_currency"`
	Line15Amount             null.Float  `json:"line_15_amount"`
	InvoiceQuoteDays         int8        `json:"invoice_quote_days"`
	InvoiceAssociateTax      null.String `json:"invoice_associate_tax"`
	InvoiceQuoteDate         time.Time   `json:"invoice_quote_date"`
	InvoiceCustomersApproval string      `json:"invoice_customers_approval"`
	Line01Notes              null.String `json:"line_01_notes"`
	Line02Notes              null.String `json:"line_02_notes"`
	TotalLabourCurrency      string      `json:"total_labour_currency"`
	TotalLabour              float64     `json:"total_labour"`
	TotalMaterialsCurrency   string      `json:"total_materials_currency"`
	TotalMaterials           float64     `json:"total_materials"`
	OtherCostsCurrency       string      `json:"other_costs_currency"`
	OtherCosts               float64     `json:"other_costs"`
	AmountDueCurrency        string      `json:"amount_due_currency"`
	TaxCurrency              string      `json:"tax_currency"`
	Tax                      float64     `json:"tax"`
	TotalCurrency            string      `json:"total_currency"`
	Total                    float64     `json:"total"`
	DepositCurrency          string      `json:"deposit_currency"`
	PaymentAmountCurrency    string      `json:"payment_amount_currency"`
	PaymentAmount            float64     `json:"payment_amount"`
	PaymentDate              time.Time   `json:"payment_date"`
	IsCash                   bool        `json:"is_cash"`
	IsCheque                 bool        `json:"is_cheque"`
	IsDebit                  bool        `json:"is_debit"`
	IsCredit                 bool        `json:"is_credit"`
	IsOther                  bool        `json:"is_other"`
	ClientSignature          string      `json:"client_signature"`
	AssociateSignDate        time.Time   `json:"associate_sign_date"`
	AssociateSignature       string      `json:"associate_signature"`
	WorkOrderID              uint64      `json:"work_order_id"`
	CreatedAt                time.Time   `json:"created_at"`
	LastModifiedAt           time.Time   `json:"last_modified_at"`
	CreatedByID              uint64      `json:"created_by_id"`
	LastModifiedByID         uint64      `json:"last_modified_by_id"`
	CreatedFrom              string      `json:"created_from"`
	CreatedFromIsPublic      bool        `json:"created_from_is_public"`
	LastModifiedFrom         string      `json:"last_modified_from"`
	LastModifiedFromIsPublic bool        `json:"last_modified_from_is_public"`
	ClientAddress            string      `json:"client_address"`
	RevisionVersion          int8        `json:"revision_version"`
	Deposit                  float64     `json:"deposit"`
	AmountDue                float64     `json:"amount_due"`
	SubTotal                 float64     `json:"sub_total"`
	SubTotalCurrency         string      `json:"sub_total_currency"`
}

func ListAllWorkOrderInvoices(db *sql.DB) ([]*OldWorkOrderInvoice, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `
	SELECT
        order_id, is_archived, invoice_id, invoice_date, associate_name,
		associate_telephone, client_name, client_telephone, client_email,
		line_01_qty, line_01_desc, line_01_price_currency, line_01_price, line_01_amount_currency, line_01_amount,
		line_02_qty, line_02_desc, line_02_price_currency, line_02_price, line_02_amount_currency, line_02_amount,
		line_03_qty, line_03_desc, line_03_price_currency, line_03_price, line_03_amount_currency, line_03_amount,
		line_04_qty, line_04_desc, line_04_price_currency, line_04_price, line_04_amount_currency, line_04_amount,
		line_05_qty, line_05_desc, line_05_price_currency, line_05_price, line_05_amount_currency, line_05_amount,
		line_06_qty, line_06_desc, line_06_price_currency, line_06_price, line_06_amount_currency, line_06_amount,
		line_07_qty, line_07_desc, line_07_price_currency, line_07_price, line_07_amount_currency, line_07_amount,
		line_08_qty, line_08_desc, line_08_price_currency, line_08_price, line_08_amount_currency, line_08_amount,
		line_09_qty, line_09_desc, line_09_price_currency, line_09_price, line_09_amount_currency, line_09_amount,
		line_10_qty, line_10_desc, line_10_price_currency, line_10_price, line_10_amount_currency, line_10_amount,
		line_11_qty, line_11_desc, line_11_price_currency, line_11_price, line_11_amount_currency, line_11_amount,
		line_12_qty, line_12_desc, line_12_price_currency, line_12_price, line_12_amount_currency, line_12_amount,
		line_13_qty, line_13_desc, line_13_price_currency, line_13_price, line_13_amount_currency, line_13_amount,
		line_14_qty, line_14_desc, line_14_price_currency, line_14_price, line_14_amount_currency, line_14_amount,
		line_15_qty, line_15_desc, line_15_price_currency, line_15_price, line_15_amount_currency, line_15_amount,
		invoice_quote_days, invoice_associate_tax, invoice_quote_date, invoice_customers_approval, line_01_notes,
		line_02_notes, total_labour_currency, total_labour, total_materials_currency, total_materials,
		other_costs_currency, other_costs, amount_due_currency, tax_currency, tax, total_currency, total,
		deposit_currency, payment_amount_currency, payment_amount, payment_date, is_cash, is_cheque, is_debit,
		is_credit, is_other, client_signature, associate_sign_date, associate_signature, work_order_id, created_at,
		last_modified_at, created_by_id, last_modified_by_id, created_from, created_from_is_public, last_modified_from,
		last_modified_from_is_public, client_address, revision_version, deposit, amount_due, sub_total, sub_total_currency
	FROM
        london.workery_work_order_invoices
	ORDER BY
	    order_id
	ASC
	`
	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}

	var arr []*OldWorkOrderInvoice
	defer rows.Close()
	for rows.Next() {
		m := new(OldWorkOrderInvoice)
		err = rows.Scan(
			&m.OrderID, &m.IsArchived, &m.InvoiceID, &m.InvoiceDate, &m.AssociateName,
			&m.AssociateTelephone, &m.ClientName, &m.ClientTelephone, &m.ClientEmail,
			&m.Line01Qty, &m.Line01Desc, &m.Line01PriceCurrency, &m.Line01Price, &m.Line01AmountCurrency, &m.Line01Amount,
			&m.Line02Qty, &m.Line02Desc, &m.Line02PriceCurrency, &m.Line02Price, &m.Line02AmountCurrency, &m.Line02Amount,
			&m.Line03Qty, &m.Line03Desc, &m.Line03PriceCurrency, &m.Line03Price, &m.Line03AmountCurrency, &m.Line03Amount,
			&m.Line04Qty, &m.Line04Desc, &m.Line04PriceCurrency, &m.Line04Price, &m.Line04AmountCurrency, &m.Line04Amount,
			&m.Line05Qty, &m.Line05Desc, &m.Line05PriceCurrency, &m.Line05Price, &m.Line05AmountCurrency, &m.Line05Amount,
			&m.Line06Qty, &m.Line06Desc, &m.Line06PriceCurrency, &m.Line06Price, &m.Line06AmountCurrency, &m.Line06Amount,
			&m.Line07Qty, &m.Line07Desc, &m.Line07PriceCurrency, &m.Line07Price, &m.Line07AmountCurrency, &m.Line07Amount,
			&m.Line08Qty, &m.Line08Desc, &m.Line08PriceCurrency, &m.Line08Price, &m.Line08AmountCurrency, &m.Line08Amount,
			&m.Line09Qty, &m.Line09Desc, &m.Line09PriceCurrency, &m.Line09Price, &m.Line09AmountCurrency, &m.Line09Amount,
			&m.Line10Qty, &m.Line10Desc, &m.Line10PriceCurrency, &m.Line10Price, &m.Line10AmountCurrency, &m.Line10Amount,
			&m.Line11Qty, &m.Line11Desc, &m.Line11PriceCurrency, &m.Line11Price, &m.Line11AmountCurrency, &m.Line11Amount,
			&m.Line12Qty, &m.Line12Desc, &m.Line12PriceCurrency, &m.Line12Price, &m.Line12AmountCurrency, &m.Line12Amount,
			&m.Line13Qty, &m.Line13Desc, &m.Line13PriceCurrency, &m.Line13Price, &m.Line13AmountCurrency, &m.Line13Amount,
			&m.Line14Qty, &m.Line14Desc, &m.Line14PriceCurrency, &m.Line14Price, &m.Line14AmountCurrency, &m.Line14Amount,
			&m.Line15Qty, &m.Line15Desc, &m.Line15PriceCurrency, &m.Line15Price, &m.Line15AmountCurrency, &m.Line15Amount,
			&m.InvoiceQuoteDays, &m.InvoiceAssociateTax, &m.InvoiceQuoteDate, &m.InvoiceCustomersApproval, &m.Line01Notes,
			&m.Line02Notes, &m.TotalLabourCurrency, &m.TotalLabour, &m.TotalMaterialsCurrency, &m.TotalMaterials,
			&m.OtherCostsCurrency, &m.OtherCosts, &m.AmountDueCurrency, &m.TaxCurrency, &m.Tax, &m.TotalCurrency, &m.Total,
			&m.DepositCurrency, &m.PaymentAmountCurrency, &m.PaymentAmount, &m.PaymentDate, &m.IsCash, &m.IsCheque,
			&m.IsDebit, &m.IsCredit, &m.IsOther, &m.ClientSignature, &m.AssociateSignDate, &m.AssociateSignature,
			&m.WorkOrderID, &m.CreatedAt, &m.LastModifiedAt, &m.CreatedByID, &m.LastModifiedByID, &m.CreatedFrom,
			&m.CreatedFromIsPublic, &m.LastModifiedFrom, &m.LastModifiedFromIsPublic, &m.ClientAddress, &m.RevisionVersion,
			&m.Deposit, &m.AmountDue, &m.SubTotal, &m.SubTotalCurrency,
		)
		if err != nil {
			log.Fatalln("rows.Scan|err:", err)
		}
		arr = append(arr, m)
	}
	err = rows.Err()
	if err != nil {
		panic(err)
	}
	return arr, err
}

func RunImportOrderInvoice(cfg *config.Conf, public *sql.DB, london *sql.DB, tenantStorer tenant_ds.TenantStorer, userStorer user_ds.UserStorer, oStorer o_ds.OrderStorer, tenant *tenant_ds.Tenant) {
	fmt.Println("Beginning importing order invoices")
	data, err := ListAllWorkOrderInvoices(london)
	if err != nil {
		log.Fatal(err)
	}

	for _, datum := range data {
		importOrderInvoice(context.Background(), tenantStorer, userStorer, oStorer, tenant, datum)
	}
	fmt.Println("Finished importing order invoices")
}

func importOrderInvoice(
	ctx context.Context,
	ts tenant_ds.TenantStorer,
	uStorer user_ds.UserStorer,
	oStorer o_ds.OrderStorer,
	tenant *tenant_ds.Tenant,
	oi *OldWorkOrderInvoice,
) {

	//
	// Lookup related.
	//

	order, err := oStorer.GetByWJID(ctx, oi.WorkOrderID)
	if err != nil {
		log.Fatal(err)
	}
	if order == nil {
		log.Fatal("order does not exist")
	}

	//
	// Get `createdByID` and `createdByName` values.
	//

	var createdByID primitive.ObjectID = primitive.NilObjectID
	var createdByName string
	createdByUser, err := uStorer.GetByOldID(ctx, oi.CreatedByID)
	if err != nil {
		log.Fatal("ur.GetByOldID", err)
	}
	if createdByUser != nil {
		createdByID = createdByUser.ID
		createdByName = createdByUser.Name
	}

	//
	// Get `modifiedByID` and `modifiedByName` values.
	//

	var modifiedByID primitive.ObjectID = primitive.NilObjectID
	var modifiedByName string
	modifiedByUser, err := uStorer.GetByOldID(ctx, oi.CreatedByID)
	if err != nil {
		log.Fatal("ur.GetByOldID", err)
	}
	if modifiedByUser != nil {
		modifiedByID = modifiedByUser.ID
		modifiedByName = modifiedByUser.Name
	}

	//
	// Create the order invoice.
	//

	oc := &o_ds.OrderInvoice{
		OrderWJID: order.WJID,
		ID:        primitive.NewObjectID(), // 1
		TenantID:  tenant.ID,               // 2
		OldID:     oi.OrderID,              // 3
		// InvoiceID: order.InvoiceID,         // 4
		OrderID:                  order.ID,                             // 5
		InvoiceDate:              oi.InvoiceDate,                       // 6
		AssociateName:            oi.AssociateName,                     // 7
		AssociateTelephone:       oi.AssociateTelephone,                // 8
		ClientName:               oi.ClientName,                        // 9
		ClientTelephone:          oi.ClientTelephone,                   // 10
		ClientEmail:              oi.ClientEmail.ValueOrZero(),         // 11
		Line01Qty:                oi.Line01Qty,                         // 12
		Line01Desc:               oi.Line01Desc,                        // 13
		Line01Price:              oi.Line01Price,                       // 14
		Line01Amount:             oi.Line01Amount,                      // 15
		Line02Qty:                oi.Line02Qty.ValueOrZero(),           // 16
		Line02Desc:               oi.Line02Desc.ValueOrZero(),          // 17
		Line02Price:              oi.Line02Price.ValueOrZero(),         // 18
		Line02Amount:             oi.Line02Amount.ValueOrZero(),        // 19
		Line03Qty:                oi.Line03Qty.ValueOrZero(),           // 20
		Line03Desc:               oi.Line03Desc.ValueOrZero(),          // 21
		Line03Price:              oi.Line03Price.ValueOrZero(),         // 22
		Line03Amount:             oi.Line03Amount.ValueOrZero(),        // 23
		Line04Qty:                oi.Line04Qty.ValueOrZero(),           // 24
		Line04Desc:               oi.Line04Desc.ValueOrZero(),          // 25
		Line04Price:              oi.Line04Price.ValueOrZero(),         // 26
		Line04Amount:             oi.Line04Amount.ValueOrZero(),        // 27
		Line05Qty:                oi.Line05Qty.ValueOrZero(),           // 28
		Line05Desc:               oi.Line05Desc.ValueOrZero(),          // 29
		Line05Price:              oi.Line05Price.ValueOrZero(),         // 30
		Line05Amount:             oi.Line05Amount.ValueOrZero(),        // 31
		Line06Qty:                oi.Line06Qty.ValueOrZero(),           // 32
		Line06Desc:               oi.Line06Desc.ValueOrZero(),          // 33
		Line06Price:              oi.Line06Price.ValueOrZero(),         // 34
		Line06Amount:             oi.Line06Amount.ValueOrZero(),        // 35
		Line07Qty:                oi.Line07Qty.ValueOrZero(),           // 36
		Line07Desc:               oi.Line07Desc.ValueOrZero(),          // 37
		Line07Price:              oi.Line07Price.ValueOrZero(),         // 38
		Line07Amount:             oi.Line07Amount.ValueOrZero(),        // 39
		Line08Qty:                oi.Line08Qty.ValueOrZero(),           // 40
		Line08Desc:               oi.Line08Desc.ValueOrZero(),          // 41
		Line08Price:              oi.Line08Price.ValueOrZero(),         // 42
		Line08Amount:             oi.Line08Amount.ValueOrZero(),        // 43
		Line09Qty:                oi.Line09Qty.ValueOrZero(),           // 44
		Line09Desc:               oi.Line09Desc.ValueOrZero(),          // 45
		Line09Price:              oi.Line09Price.ValueOrZero(),         // 46
		Line09Amount:             oi.Line09Amount.ValueOrZero(),        // 47
		Line10Qty:                oi.Line10Qty.ValueOrZero(),           // 48
		Line10Desc:               oi.Line10Desc.ValueOrZero(),          // 49
		Line10Price:              oi.Line10Price.ValueOrZero(),         // 50
		Line10Amount:             oi.Line10Amount.ValueOrZero(),        // 51
		Line11Qty:                oi.Line11Qty.ValueOrZero(),           // 52
		Line11Desc:               oi.Line11Desc.ValueOrZero(),          // 53
		Line11Price:              oi.Line11Price.ValueOrZero(),         // 54
		Line11Amount:             oi.Line11Amount.ValueOrZero(),        // 55
		Line12Qty:                oi.Line12Qty.ValueOrZero(),           // 56
		Line12Desc:               oi.Line12Desc.ValueOrZero(),          // 57
		Line12Price:              oi.Line12Price.ValueOrZero(),         // 58
		Line12Amount:             oi.Line12Amount.ValueOrZero(),        // 59
		Line13Qty:                oi.Line13Qty.ValueOrZero(),           // 60
		Line13Desc:               oi.Line13Desc.ValueOrZero(),          // 61
		Line13Price:              oi.Line13Price.ValueOrZero(),         // 62
		Line13Amount:             oi.Line13Amount.ValueOrZero(),        // 63
		Line14Qty:                oi.Line14Qty.ValueOrZero(),           // 64
		Line14Desc:               oi.Line14Desc.ValueOrZero(),          // 65
		Line14Price:              oi.Line14Price.ValueOrZero(),         // 66
		Line14Amount:             oi.Line14Amount.ValueOrZero(),        // 67
		Line15Qty:                oi.Line15Qty.ValueOrZero(),           // 68
		Line15Desc:               oi.Line15Desc.ValueOrZero(),          // 69
		Line15Price:              oi.Line15Price.ValueOrZero(),         // 70
		Line15Amount:             oi.Line15Amount.ValueOrZero(),        // 71
		InvoiceQuoteDays:         oi.InvoiceQuoteDays,                  // 72
		InvoiceAssociateTax:      oi.InvoiceAssociateTax.ValueOrZero(), // 73
		InvoiceQuoteDate:         oi.InvoiceQuoteDate,                  // 74
		InvoiceCustomersApproval: oi.InvoiceCustomersApproval,          // 75
		Line01Notes:              oi.Line01Notes.ValueOrZero(),         // 76
		Line02Notes:              oi.Line02Notes.ValueOrZero(),         // 77
		TotalLabour:              oi.TotalLabour,                       // 78
		TotalMaterials:           oi.TotalMaterials,                    // 79
		OtherCosts:               oi.OtherCosts,                        // 80
		Tax:                      oi.Tax,                               // 81
		Total:                    oi.Total,                             // 82
		PaymentAmount:            oi.PaymentAmount,                     // 83
		PaymentDate:              oi.PaymentDate,                       // 84
		IsCash:                   oi.IsCash,                            // 85
		IsCheque:                 oi.IsCheque,                          // 86
		IsDebit:                  oi.IsDebit,                           // 87
		IsCredit:                 oi.IsCredit,                          // 88
		IsOther:                  oi.IsOther,                           // 89
		ClientSignature:          oi.ClientSignature,                   // 90
		AssociateSignDate:        oi.AssociateSignDate,                 // 91
		AssociateSignature:       oi.AssociateSignature,                // 92
		// WorkOrderId            uint64 `json:"work_order_id"`
		CreatedAt:          oi.CreatedAt,      // 93
		ModifiedAt:         oi.LastModifiedAt, // 94
		CreatedByUserID:    createdByID,       // 95
		CreatedByUserName:  createdByName,
		ModifiedByUserID:   modifiedByID, // 96
		ModifiedByUserName: modifiedByName,
		// CreatedFrom:         oi.CreatedFrom,
		// CreatedFromIsPublic bool `json:"created_from_is_public"`
		// LastModifiedFrom string `json:"last_modified_from"`
		// LastModifiedFromIsPublic bool `json:"last_modified_from_is_public"`
		ClientAddress:   oi.ClientAddress,   // 97
		RevisionVersion: oi.RevisionVersion, // 98
		Deposit:         oi.Deposit,         // 99
		AmountDue:       oi.AmountDue,       // 100
		SubTotal:        oi.SubTotal,        // 101
		// State:                 oi.State,           // 102 (TODO)
	}

	// Append invoices to order details.
	order.Invoices = append(order.Invoices, oc)

	if err := oStorer.UpdateByID(ctx, order); err != nil {
		log.Fatal(err)
	}

	//
	// For debugging purposes only.
	//

	fmt.Println("Imported Order Invoice ID#", oc.ID, "for OrderID", order.ID)
}
