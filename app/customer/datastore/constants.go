package datastore

const (
	CustomerStatusActive                    = 1
	CustomerStatusArchived                  = 2
	CustomerDeactivationReasonNotSpecified  = 0
	CustomerDeactivationReasonOther         = 1
	CustomerDeactivationReasonBlacklisted   = 2
	CustomerDeactivationReasonMoved         = 3
	CustomerDeactivationReasonDeceased      = 4
	CustomerDeactivationReasonDoNotConstact = 5
	CustomerTypeUnassigned                  = 1
	CustomerTypeResidential                 = 2
	CustomerTypeCommercial                  = 3
	CustomerPhoneTypeLandline               = 1
	CustomerPhoneTypeMobile                 = 2
	CustomerPhoneTypeWork                   = 3
	CustomerOrganizationTypeUnknown         = 1
	CustomerOrganizationTypePrivate         = 2
	CustomerOrganizationTypeNonProfit       = 3
	CustomerOrganizationTypeGovernment      = 4
	CustomerGenderOther                     = 1
	CustomerGenderMan                       = 2
	CustomerGenderWoman                     = 3
	CustomerGenderTransgender               = 4
	CustomerGenderNonBinary                 = 5
	CustomerGenderTwoSpirit                 = 6
	CustomerGenderPreferNotToSay            = 7
	CustomerGenderDoNotKnow                 = 8
)

var CustomerStateLabels = map[int8]string{
	CustomerStatusActive:   "Active",
	CustomerStatusArchived: "Archived",
}

var CustomerTypeLabels = map[int8]string{
	CustomerTypeResidential: "Residential",
	CustomerTypeCommercial:  "Commercial",
	CustomerTypeUnassigned:  "Unassigned",
}

var CustomerDeactivationReasonLabels = map[int8]string{
	CustomerDeactivationReasonNotSpecified:  "Not Specified",
	CustomerDeactivationReasonOther:         "Other",
	CustomerDeactivationReasonBlacklisted:   "Blacklisted",
	CustomerDeactivationReasonMoved:         "Moved",
	CustomerDeactivationReasonDeceased:      "Deceased",
	CustomerDeactivationReasonDoNotConstact: "Do not contact",
}

var CustomerTelephoneTypeLabels = map[int8]string{
	1: "Landline",
	2: "Mobile",
	3: "Work",
}

var CustomerOrganizationTypeLabels = map[int8]string{
	1: "Unknown",
	2: "Private",
	3: "Non-Profit",
	4: "Government",
}

var CustomerGenderLabels = map[int8]string{
	CustomerGenderOther:          "Other",
	CustomerGenderMan:            "Man",
	CustomerGenderWoman:          "Women",
	CustomerGenderNonBinary:      "Non-Binary",
	CustomerGenderTwoSpirit:      "Two Spirit",
	CustomerGenderPreferNotToSay: "Prefer Not To Say",
	CustomerGenderDoNotKnow:      "Do Not Know",
}
