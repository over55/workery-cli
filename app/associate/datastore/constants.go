package datastore

const (
	AssociateStatusActive   = 1
	AssociateStatusArchived = 2

	AssociateDeactivationReasonNotSpecified  = 0
	AssociateDeactivationReasonOther         = 1
	AssociateDeactivationReasonBlacklisted   = 2
	AssociateDeactivationReasonMoved         = 3
	AssociateDeactivationReasonDeceased      = 4
	AssociateDeactivationReasonDoNotConstact = 5

	AssociateTypeUnassigned  = 1
	AssociateTypeResidential = 2
	AssociateTypeCommercial  = 3

	AssociatePhoneTypeLandline = 1
	AssociatePhoneTypeMobile   = 2
	AssociatePhoneTypeWork     = 3

	AssociateGenderOther          = 1
	AssociateGenderMan            = 2
	AssociateGenderWoman          = 3
	AssociateGenderTransgender    = 4
	AssociateGenderNonBinary      = 5
	AssociateGenderTwoSpirit      = 6
	AssociateGenderPreferNotToSay = 7
	AssociateGenderDoNotKnow      = 8

	AssociateIdentifyAsOther                = 1
	AssociateIdentifyAsPreferNotToSay       = 2
	AssociateIdentifyAsWomen                = 3
	AssociateIdentifyAsNewcomer             = 4
	AssociateIdentifyAsRacializedPerson     = 5
	AssociateIdentifyAsVeteran              = 6
	AssociateIdentifyAsFrancophone          = 7
	AssociateIdentifyAsPersonWithDisability = 8
	AssociateIdentifyAsInuit                = 9
	AssociateIdentifyAsFirstNations         = 10
	AssociateIdentifyAsMetis                = 11
)

var AssociateStateLabels = map[int8]string{
	AssociateStatusActive:   "Active",
	AssociateStatusArchived: "Archived",
}

var AssociateTypeLabels = map[int8]string{
	AssociateTypeResidential: "Residential",
	AssociateTypeCommercial:  "Commercial",
	AssociateTypeUnassigned:  "Unassigned",
}

var AssociateDeactivationReasonLabels = map[int8]string{
	AssociateDeactivationReasonNotSpecified:  "Not Specified",
	AssociateDeactivationReasonOther:         "Other",
	AssociateDeactivationReasonBlacklisted:   "Blacklisted",
	AssociateDeactivationReasonMoved:         "Moved",
	AssociateDeactivationReasonDeceased:      "Deceased",
	AssociateDeactivationReasonDoNotConstact: "Do not contact",
}

var AssociateTelephoneTypeLabels = map[int8]string{
	1: "Landline",
	2: "Mobile",
	3: "Work",
}

var AssociateOrganizationTypeLabels = map[int8]string{
	1: "Unknown",
	2: "Private",
	3: "Non-Profit",
	4: "Government",
}

var AssociateGenderLabels = map[int8]string{
	AssociateGenderOther:          "Other",
	AssociateGenderMan:            "Man",
	AssociateGenderWoman:          "Women",
	AssociateGenderNonBinary:      "Non-Binary",
	AssociateGenderTwoSpirit:      "Two Spirit",
	AssociateGenderPreferNotToSay: "Prefer Not To Say",
	AssociateGenderDoNotKnow:      "Do Not Know",
}
