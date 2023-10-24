package datastore

const (
	OrderStatusActive   = 1
	OrderStatusArchived = 2
)

const (
	OrderArchivedState             = 0
	OrderNewState                  = 1
	OrderDeclinedState             = 2
	OrderPendingState              = 3
	OrderCancelledState            = 4
	OrderOngoingState              = 5
	OrderInProgressState           = 6
	OrderCompletedButUnpaidState   = 7
	OrderCompletedAndPaidState     = 8
	OrderAnotherUnassignedType     = 0
	OrderResidentialType           = 1
	OrderCommercialType            = 2
	OrderUnassignedType            = 3
	OrderAssociateInvoicePaidTo    = 1
	OrderOrganizationInvoicePaidTo = 2
)

var OrderOrganizationInvoicePaidToLabels = map[int8]string{
	OrderAssociateInvoicePaidTo:    "Associate",
	OrderOrganizationInvoicePaidTo: "Organization",
}

var OrderTypeLabels = map[int8]string{
	OrderResidentialType:       "Residential",
	OrderCommercialType:        "Commercial",
	OrderUnassignedType:        "Unassigned",
	OrderAnotherUnassignedType: "-",
}

var OrderStateLabels = map[int8]string{
	OrderArchivedState:           "Archived",
	OrderNewState:                "New",
	OrderDeclinedState:           "Declined",
	OrderPendingState:            "Pending",
	OrderCancelledState:          "Cancelled",
	OrderOngoingState:            "Ongoing",
	OrderInProgressState:         "In Progress",
	OrderCompletedButUnpaidState: "Completed but Unpaid",
	OrderCompletedAndPaidState:   "Completed and Paid",
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
