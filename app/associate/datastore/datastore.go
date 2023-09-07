package datastore

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/exp/slog"

	c "github.com/over55/workery-cli/config"
)

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
)

type Associate struct {
	ID                                   primitive.ObjectID               `bson:"_id" json:"id"`
	TenantID                             primitive.ObjectID               `bson:"tenant_id" json:"tenant_id,omitempty"`
	FirstName                            string                           `bson:"first_name" json:"first_name"`
	LastName                             string                           `bson:"last_name" json:"last_name"`
	Name                                 string                           `bson:"name" json:"name"`
	LexicalName                          string                           `bson:"lexical_name" json:"lexical_name"`
	Email                                string                           `bson:"email" json:"email"`
	IsOkToEmail                          bool                             `bson:"is_ok_to_email" json:"is_ok_to_email"`
	Phone                                string                           `bson:"phone" json:"phone,omitempty"`
	PhoneType                            int8                             `bson:"phone_type" json:"phone_type"`
	PhoneExtension                       string                           `bson:"phone_extension" json:"phone_extension"`
	IsOkToText                           bool                             `bson:"is_ok_to_text" json:"is_ok_to_text"`
	FaxNumber                            string                           `bson:"fax_number" json:"fax_number"`
	OtherPhone                           string                           `bson:"other_phone" json:"other_phone"`
	OtherPhoneExtension                  string                           `bson:"other_phone_extension" json:"other_phone_extension"`
	OtherPhoneType                       int8                             `bson:"other_phone_type" json:"other_phone_type"`
	Country                              string                           `bson:"country" json:"country,omitempty"`
	Region                               string                           `bson:"region" json:"region,omitempty"`
	City                                 string                           `bson:"city" json:"city,omitempty"`
	PostalCode                           string                           `bson:"postal_code" json:"postal_code,omitempty"`
	AddressLine1                         string                           `bson:"address_line1" json:"address_line1,omitempty"`
	AddressLine2                         string                           `bson:"address_line2" json:"address_line2,omitempty"`
	PostOfficeBoxNumber                  string                           `bson:"post_office_box_number" json:"post_office_box_number"`
	FullAddressWithoutPostalCode         string                           `bson:"full_address_without_postal_code" json:"full_address_without_postal_code,omitempty"` // Compiled value
	FullAddressWithPostalCode            string                           `bson:"full_address_with_postal_code" json:"full_address_with_postal_code,omitempty"`       // Compiled value
	FullAddressURL                       string                           `bson:"full_address_url" json:"full_address_url,omitempty"`                                 // Compiled value
	HasShippingAddress                   bool                             `bson:"has_shipping_address" json:"has_shipping_address,omitempty"`
	ShippingName                         string                           `bson:"shipping_name" json:"shipping_name,omitempty"`
	ShippingPhone                        string                           `bson:"shipping_phone" json:"shipping_phone,omitempty"`
	ShippingCountry                      string                           `bson:"shipping_country" json:"shipping_country,omitempty"`
	ShippingRegion                       string                           `bson:"shipping_region" json:"shipping_region,omitempty"`
	ShippingCity                         string                           `bson:"shipping_city" json:"shipping_city,omitempty"`
	ShippingPostalCode                   string                           `bson:"shipping_postal_code" json:"shipping_postal_code,omitempty"`
	ShippingAddressLine1                 string                           `bson:"shipping_address_line1" json:"shipping_address_line1,omitempty"`
	ShippingAddressLine2                 string                           `bson:"shipping_address_line2" json:"shipping_address_line2,omitempty"`
	ShippingPostOfficeBoxNumber          string                           `bson:"shipping_post_office_box_number" json:"shipping_post_office_box_number"`
	ShippingFullAddressWithoutPostalCode string                           `bson:"shipping_full_address_without_postal_code" json:"shipping_full_address_without_postal_code,omitempty"` // Compiled value
	ShippingFullAddressWithPostalCode    string                           `bson:"shipping_full_address_with_postal_code" json:"shipping_full_address_with_postal_code,omitempty"`       // Compiled value
	ShippingFullAddressURL               string                           `bson:"shipping_full_address_url" json:"shipping_full_address_url,omitempty"`                                 // Compiled value
	HowDidYouHearAboutUsID               primitive.ObjectID               `bson:"how_did_you_hear_about_us_id" json:"how_did_you_hear_about_us_id,omitempty"`
	HowDidYouHearAboutUsText             string                           `bson:"how_did_you_hear_about_us_text" json:"how_did_you_hear_about_us_text,omitempty"`
	IsHowDidYouHearAboutUsOther          bool                             `bson:"is_how_did_you_hear_about_us_other" json:"is_how_did_you_hear_about_us_other,omitempty"`
	HowDidYouHearAboutUsOther            string                           `bson:"how_did_you_hear_about_us_other" json:"how_did_you_hear_about_us_other,omitempty"`
	AgreeTOS                             bool                             `bson:"agree_tos" json:"agree_tos,omitempty"`
	CreatedAt                            time.Time                        `bson:"created_at" json:"created_at,omitempty"`
	CreatedByUserID                      primitive.ObjectID               `bson:"created_by_user_id" json:"created_by_user_id"`
	CreatedByUserName                    string                           `bson:"created_by_user_name" json:"created_by_user_name"`
	CreatedFromIPAddress                 string                           `bson:"created_from_ip_address" json:"created_from_ip_address"`
	ModifiedAt                           time.Time                        `bson:"modified_at" json:"modified_at,omitempty"`
	ModifiedByUserID                     primitive.ObjectID               `bson:"modified_by_user_id" json:"modified_by_user_id"`
	ModifiedByUserName                   string                           `bson:"modified_by_user_name" json:"modified_by_user_name"`
	ModifiedFromIPAddress                string                           `bson:"modified_from_ip_address" json:"modified_from_ip_address"`
	Status                               int8                             `bson:"status" json:"status"`
	Salt                                 string                           `bson:"salt" json:"salt,omitempty"`
	PrAccessCode                         string                           `bson:"pr_access_code" json:"pr_access_code,omitempty"`
	PrExpiryTime                         time.Time                        `bson:"pr_expiry_time" json:"pr_expiry_time,omitempty"`
	Timezone                             string                           `bson:"timezone" json:"timezone,omitempty"`
	HasUserAccount                       bool                             `bson:"has_user_account" json:"has_user_account,omitempty"`
	UserID                               primitive.ObjectID               `bson:"user_id" json:"user_id,omitempty"`
	Type                                 int8                             `bson:"type" json:"type"`
	IsBusiness                           bool                             `bson:"is_business" json:"is_business"`
	IsSenior                             bool                             `bson:"is_senior" json:"is_senior"`
	IsSupport                            bool                             `bson:"is_support" json:"is_support"`
	JobInfoRead                          string                           `bson:"job_info_read" json:"job_info_read"`
	DeactivationReason                   int8                             `bson:"deactivation_reason" json:"deactivation_reason"`
	DeactivationReasonOther              string                           `bson:"deactivation_reason_other" json:"deactivation_reason_other"`
	Description                          string                           `bson:"description" json:"description"`
	AvatarObjectExpiry                   time.Time                        `bson:"avatar_object_expiry" json:"avatar_object_expiry"`
	AvatarObjectURL                      string                           `bson:"avatar_object_url" json:"avatar_object_url"`
	AvatarObjectKey                      string                           `bson:"avatar_object_key" json:"avatar_object_key"`
	AvatarFileType                       string                           `bson:"avatar_file_type" json:"avatar_file_type"`
	AvatarFileName                       string                           `bson:"avatar_file_name" json:"avatar_file_name"`
	BirthDate                            time.Time                        `bson:"birth_date" json:"birth_date"`
	JoinDate                             time.Time                        `bson:"join_date" json:"join_date"`
	Nationality                          string                           `bson:"nationality" json:"nationality"`
	Gender                               string                           `bson:"gender" json:"gender"`
	TaxID                                string                           `bson:"tax_id" json:"tax_id"`
	Elevation                            float64                          `bson:"elevation" json:"elevation"`
	Latitude                             float64                          `bson:"latitude" json:"latitude"`
	Longitude                            float64                          `bson:"longitude" json:"longitude"`
	AreaServed                           string                           `bson:"area_served" json:"area_served"`
	AvailableLanguage                    string                           `bson:"available_language" json:"available_language"`
	ContactType                          string                           `bson:"contact_type" json:"contact_type"`
	OrganizationName                     string                           `bson:"organization_name" json:"organization_name"`
	OrganizationType                     int8                             `bson:"organization_type" json:"organization_type"`
	OldID                                uint64                           `bson:"old_id" json:"old_id,omitempty"`
	HourlySalaryDesired                  int64                            `bson:"hourly_salary_desired" json:"hourly_salary_desired"`
	LimitSpecial                         string                           `bson:"limit_special" json:"limit_special"`
	DuesDate                             time.Time                        `bson:"dues_date" json:"dues_date"`
	CommercialInsuranceExpiryDate        time.Time                        `bson:"commercial_insurance_expiry_date" json:"commercial_insurance_expiry_date"`
	AutoInsuranceExpiryDate              time.Time                        `bson:"auto_insurance_expiry_date" json:"auto_insurance_expiry_date"`
	WsibNumber                           string                           `bson:"wsib_number" json:"wsib_number"`
	WsibInsuranceDate                    time.Time                        `bson:"wsib_insurance_date" json:"wsib_insurance_date"`
	PoliceCheck                          time.Time                        `bson:"police_check" json:"police_check"`
	DriversLicenseClass                  string                           `bson:"drivers_license_class" json:"drivers_license_class"`
	Score                                float64                          `bson:"score" json:"score"`
	ServiceFeeID                         primitive.ObjectID               `bson:"service_fee_id" json:"service_fee_id"`
	ServiceFeeName                       string                           `bson:"service_fee_name" json:"service_fee_name"`
	ServiceFeePercentage                 float64                          `bson:"service_fee_percentage" json:"service_fee_percentage"`
	BalanceOwingAmount                   float64                          `bson:"balance_owing_amount" json:"balance_owing_amount"`
	EmergencyContactName                 string                           `bson:"emergency_contact_name" json:"emergency_contact_name"`
	EmergencyContactRelationship         string                           `bson:"emergency_contact_relationship" json:"emergency_contact_relationship"`
	EmergencyContactTelephone            string                           `bson:"emergency_contact_telephone" json:"emergency_contact_telephone"`
	EmergencyContactAlternativeTelephone string                           `bson:"emergency_contact_alternative_telephone" json:"emergency_contact_alternative_telephone"`
	Comments                             []*AssociateComment              `bson:"comments" json:"comments"`
	SkillSets                            []*AssociateSkillSet             `bson:"skill_sets" json:"skill_sets,omitempty"`
	InsuranceRequirements                []*AssociateInsuranceRequirement `bson:"insurance_requirements" json:"insurance_requirements,omitempty"`
	VehicleTypes                         []*AssociateVehicleType          `bson:"vehicle_types" json:"vehicle_types,omitempty"`
	AwayLogs                             []*AssociateAwayLog              `bson:"away_logs" json:"away_logs,omitempty"`
	Tags                                 []*AssociateTag                  `bson:"tags" json:"tags,omitempty"`
	// ServiceFee            *WorkOrderServiceFee             `json:"invoice_service_fee,omitempty"` // Referenced value from 'work_order_service_fee'.
	// Tags                  []*AssociateTag                  `json:"tags,omitempty"`
}

type AssociateComment struct {
	ID                    primitive.ObjectID `bson:"_id" json:"id"`
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
	OldID                 uint64             `bson:"old_id" json:"old_id"`
}

type AssociateVehicleType struct {
	ID          primitive.ObjectID `bson:"_id" json:"id"`
	TenantID    primitive.ObjectID `bson:"tenant_id" json:"tenant_id,omitempty"`
	Name        string             `bson:"name" json:"name"`
	Description string             `bson:"description" json:"description"`
	Status      int8               `bson:"status" json:"status"`
	OldID       uint64             `bson:"old_id" json:"old_id"`
}

type AssociateSkillSet struct {
	ID          primitive.ObjectID `bson:"_id" json:"id"`
	TenantID    primitive.ObjectID `bson:"tenant_id" json:"tenant_id,omitempty"`
	Category    string             `bson:"category" json:"category"`
	SubCategory string             `bson:"sub_category" json:"sub_category"`
	Description string             `bson:"description" json:"description"`
	Status      int8               `bson:"status" json:"status"`
	OldID       uint64             `bson:"old_id" json:"old_id"`
}

type AssociateInsuranceRequirement struct {
	ID          primitive.ObjectID `bson:"_id" json:"id"`
	TenantID    primitive.ObjectID `bson:"tenant_id" json:"tenant_id,omitempty"`
	Name        string             `bson:"name" json:"name"`
	Description string             `bson:"description" json:"description"`
	Status      int8               `bson:"status" json:"status"`
	OldID       uint64             `bson:"old_id" json:"old_id"`
}

type AssociateAwayLog struct {
	ID                    primitive.ObjectID `bson:"_id" json:"id"`
	TenantID              primitive.ObjectID `bson:"tenant_id" json:"tenant_id,omitempty"`
	AssociateID           primitive.ObjectID `bson:"associate_id" json:"associate_id"`
	AssociateName         string             `bson:"associate_name" json:"associate_name,omitempty"`
	AssociateLexicalName  string             `bson:"associate_lexical_name" json:"associate_lexical_name,omitempty"`
	Reason                int8               `bson:"reason" json:"reason"`
	ReasonOther           string             `bson:"reason_other" json:"reason_other"`
	UntilFurtherNotice    bool               `bson:"until_further_notice" json:"until_further_notice"`
	UntilDate             time.Time          `bson:"until_date" json:"until_date"`
	StartDate             time.Time          `bson:"start_date" json:"start_date"`
	Status                int8               `bson:"status" json:"status"`
	CreatedAt             time.Time          `bson:"created_at" json:"created_at"`
	CreatedByUserID       primitive.ObjectID `bson:"created_by_user_id" json:"created_by_user_id,omitempty"`
	CreatedByUserName     string             `bson:"created_by_user_name" json:"created_by_user_name"`
	CreatedFromIPAddress  string             `bson:"created_from_ip_address" json:"created_from_ip_address"`
	ModifiedAt            time.Time          `bson:"modified_at" json:"modified_at"`
	ModifiedByUserID      primitive.ObjectID `bson:"modified_by_user_id" json:"modified_by_user_id,omitempty"`
	ModifiedByUserName    string             `bson:"modified_by_user_name" json:"modified_by_user_name"`
	ModifiedFromIPAddress string             `bson:"modified_from_ip_address" json:"modified_from_ip_address"`
	OldID                 uint64             `bson:"old_id" json:"old_id"`
}

type AssociateTag struct {
	ID          primitive.ObjectID `bson:"_id" json:"id"`
	TenantID    primitive.ObjectID `bson:"tenant_id" json:"tenant_id,omitempty"`
	Text        string             `bson:"text" json:"text"`
	Description string             `bson:"description" json:"description"`
	Status      int8               `bson:"status" json:"status"`
	OldID       uint64             `bson:"old_id" json:"old_id"`
}

type AssociateListFilter struct {
	// Pagination related.
	Cursor    primitive.ObjectID
	PageSize  int64
	SortField string
	SortOrder int8 // 1=ascending | -1=descending

	// Filter related.
	TenantID        primitive.ObjectID
	Role            int8
	Status          int8
	UUIDs           []string
	ExcludeArchived bool
	SearchText      string
	FirstName       string
	LastName        string
	Email           string
	Phone           string
	CreatedAtGTE    time.Time
}

type AssociateListResult struct {
	Results     []*Associate       `json:"results"`
	NextCursor  primitive.ObjectID `json:"next_cursor"`
	HasNextPage bool               `json:"has_next_page"`
}

type AssociateAsSelectOption struct {
	Value primitive.ObjectID `bson:"_id" json:"value"` // Extract from the database `_id` field and output through API as `value`.
	Label string             `bson:"name" json:"label"`
}

// AssociateStorer Interface for user.
type AssociateStorer interface {
	Create(ctx context.Context, m *Associate) error
	GetByID(ctx context.Context, id primitive.ObjectID) (*Associate, error)
	GetByOldID(ctx context.Context, oldID uint64) (*Associate, error)
	GetByEmail(ctx context.Context, email string) (*Associate, error)
	GetByVerificationCode(ctx context.Context, verificationCode string) (*Associate, error)
	CheckIfExistsByEmail(ctx context.Context, email string) (bool, error)
	UpdateByID(ctx context.Context, m *Associate) error
	UpsertByID(ctx context.Context, user *Associate) error
	ListByFilter(ctx context.Context, f *AssociateListFilter) (*AssociateListResult, error)
	ListAsSelectOptionByFilter(ctx context.Context, f *AssociateListFilter) ([]*AssociateAsSelectOption, error)
	DeleteByID(ctx context.Context, id primitive.ObjectID) error
	// //TODO: Add more...
}

type AssociateStorerImpl struct {
	Logger     *slog.Logger
	DbClient   *mongo.Client
	Collection *mongo.Collection
}

func NewDatastore(appCfg *c.Conf, loggerp *slog.Logger, client *mongo.Client) AssociateStorer {
	// ctx := context.Background()
	uc := client.Database(appCfg.DB.Name).Collection("associates")

	// The following few lines of code will create the index for our app for this
	// colleciton.
	indexModel := mongo.IndexModel{
		Keys: bson.D{
			{"name", "text"},
			{"lexical_name", "text"},
			{"email", "text"},
			{"phone", "text"},
			{"country", "text"},
			{"region", "text"},
			{"city", "text"},
			{"postal_code", "text"},
			{"address_line1", "text"},
		},
	}
	_, err := uc.Indexes().CreateOne(context.TODO(), indexModel)
	if err != nil {
		// It is important that we crash the app on startup to meet the
		// requirements of `google/wire` framework.
		log.Fatal(err)
	}

	s := &AssociateStorerImpl{
		Logger:     loggerp,
		DbClient:   client,
		Collection: uc,
	}
	return s
}

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

//---------------------
// organization_type
//---------------------
// 1 = Unknown Organization Type | UNKNOWN_ORGANIZATION_TYPE_OF_ID
// 2 = Private Organization Type | PRIVATE_ORGANIZATION_TYPE_OF_ID
// 3 = Non-Profit Organization Type | NON_PROFIT_ORGANIZATION_TYPE_OF_ID
// 4 = Government Organization | GOVERNMENT_ORGANIZATION_TYPE_OF_ID
