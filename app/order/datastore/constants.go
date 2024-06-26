package datastore

const (
	OrderStatusNew                = 1
	OrderStatusDeclined           = 2
	OrderStatusPending            = 3
	OrderStatusCancelled          = 4
	OrderStatusOngoing            = 5
	OrderStatusInProgress         = 6
	OrderStatusCompletedButUnpaid = 7
	OrderStatusCompletedAndPaid   = 8
	OrderStatusArchived           = 9

	OrderTypeAnotherUnassigned = 0
	OrderTypeResidential       = 1
	OrderTypeCommercial        = 2
	OrderTypeUnassigned        = 3

	OrderInvoicePaidToAssociate    = 1
	OrderInvoicePaidToOrganization = 2

	OrderUnassignedReasonOther                        = 1
	OrderUnassignedReasonAossicateNotAFit             = 2
	OrderUnassignedReasonJobBiggerThanThought         = 3
	OrderUnassignedReasonClientChangedJobRequirements = 4
	OrderUnassignedReasonAssociateNeedsMoreTime       = 5

	PaymentMethodOther          = 1
	PaymentMethodCash           = 2
	PaymentMethodCheque         = 3
	PaymentMethodETransfer      = 4
	PaymentMethodDebit          = 5
	PaymentMethodCredit         = 6
	PaymentMethodPurchaseOrder  = 7
	PaymentMethodCryptocurrency = 8
)

var OrderUnassignedReasonToLabels = map[int8]string{
	OrderUnassignedReasonOther:                        "Other",
	OrderUnassignedReasonAossicateNotAFit:             "Associate not a fit",
	OrderUnassignedReasonJobBiggerThanThought:         "Job bigger than thought",
	OrderUnassignedReasonClientChangedJobRequirements: "Client changed job requirements",
	OrderUnassignedReasonAssociateNeedsMoreTime:       "Associate needs more time",
}

var OrderOrganizationInvoicePaidToLabels = map[int8]string{
	OrderInvoicePaidToAssociate:    "Associate",
	OrderInvoicePaidToOrganization: "Organization",
}

var OrderTypeLabels = map[int8]string{
	OrderTypeResidential:       "Residential",
	OrderTypeCommercial:        "Commercial",
	OrderTypeUnassigned:        "Unassigned",
	OrderTypeAnotherUnassigned: "-",
}

var OrderStatusLabels = map[int8]string{
	OrderStatusArchived:           "Archived",
	OrderStatusNew:                "New",
	OrderStatusDeclined:           "Declined",
	OrderStatusPending:            "Pending",
	OrderStatusCancelled:          "Cancelled",
	OrderStatusOngoing:            "Ongoing",
	OrderStatusInProgress:         "In Progress",
	OrderStatusCompletedButUnpaid: "Completed but Unpaid",
	OrderStatusCompletedAndPaid:   "Completed and Paid",
}

var OrderClosingReasonLabels = map[int8]string{
	1:  "Other",
	2:  "Quote was too high",
	3:  "Job completed by someone else",
	4:  "Unspecified",
	5:  "Work no longer needed",
	6:  "Client not satisfied with Associate",
	7:  "Client did work themselves",
	8:  "No Associate available",
	9:  "Work environment unsuitable",
	10: "Client did not return call",
	11: "Associate did not have necessary equipment",
	12: "Repair not possible",
	13: "Could not meet deadline",
	14: "Associate did not call client",
	15: "Member issue",
	16: "Client billing issue",
}

var PaymentMethodLabels = map[int8]string{
	PaymentMethodOther:          "Other",
	PaymentMethodCash:           "Cash",
	PaymentMethodCheque:         "Cheque",
	PaymentMethodETransfer:      "E-Transfer",
	PaymentMethodDebit:          "Debit",
	PaymentMethodCredit:         "Credit",
	PaymentMethodPurchaseOrder:  "Purchase Order",
	PaymentMethodCryptocurrency: "Cryptocurrency",
}
