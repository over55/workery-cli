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
	StaffStatusActive   = 1
	StaffStatusArchived = 2

	StaffDeactivationReasonNotSpecified  = 0
	StaffDeactivationReasonOther         = 1
	StaffDeactivationReasonBlacklisted   = 2
	StaffDeactivationReasonMoved         = 3
	StaffDeactivationReasonDeceased      = 4
	StaffDeactivationReasonDoNotConstact = 5

	StaffTypeUnassigned  = 1
	StaffTypeResidential = 2
	StaffTypeCommercial  = 3

	StaffPhoneTypeLandline = 1
	StaffPhoneTypeMobile   = 2
	StaffPhoneTypeWork     = 3
)

type Staff struct {
	ID                                   primitive.ObjectID           `bson:"_id" json:"id"`
	TenantID                             primitive.ObjectID           `bson:"tenant_id" json:"tenant_id,omitempty"`
	FirstName                            string                       `bson:"first_name" json:"first_name"`
	LastName                             string                       `bson:"last_name" json:"last_name"`
	Name                                 string                       `bson:"name" json:"name"`
	LexicalName                          string                       `bson:"lexical_name" json:"lexical_name"`
	Email                                string                       `bson:"email" json:"email"`
	PersonalEmail                        string                       `bson:"personal_email" json:"personal_email"`
	Phone                                string                       `bson:"phone" json:"phone,omitempty"`
	PhoneType                            int8                         `bson:"phone_type" json:"phone_type"`
	PhoneExtension                       string                       `bson:"phone_extension" json:"phone_extension"`
	FaxNumber                            string                       `bson:"fax_number" json:"fax_number"`
	OtherPhone                           string                       `bson:"other_phone" json:"other_phone"`
	OtherPhoneExtension                  string                       `bson:"other_phone_extension" json:"other_phone_extension"`
	OtherPhoneType                       int8                         `bson:"other_phone_type" json:"other_phone_type"`
	Country                              string                       `bson:"country" json:"country,omitempty"`
	Region                               string                       `bson:"region" json:"region,omitempty"`
	City                                 string                       `bson:"city" json:"city,omitempty"`
	PostalCode                           string                       `bson:"postal_code" json:"postal_code,omitempty"`
	AddressLine1                         string                       `bson:"address_line1" json:"address_line1,omitempty"`
	AddressLine2                         string                       `bson:"address_line2" json:"address_line2,omitempty"`
	PostOfficeBoxNumber                  string                       `bson:"post_office_box_number" json:"post_office_box_number"`
	FullAddressWithoutPostalCode         string                       `bson:"full_address_without_postal_code" json:"full_address_without_postal_code,omitempty"` // Compiled value
	FullAddressWithPostalCode            string                       `bson:"full_address_with_postal_code" json:"full_address_with_postal_code,omitempty"`       // Compiled value
	FullAddressURL                       string                       `bson:"full_address_url" json:"full_address_url,omitempty"`                                 // Compiled value
	HasShippingAddress                   bool                         `bson:"has_shipping_address" json:"has_shipping_address,omitempty"`
	ShippingName                         string                       `bson:"shipping_name" json:"shipping_name,omitempty"`
	ShippingPhone                        string                       `bson:"shipping_phone" json:"shipping_phone,omitempty"`
	ShippingCountry                      string                       `bson:"shipping_country" json:"shipping_country,omitempty"`
	ShippingRegion                       string                       `bson:"shipping_region" json:"shipping_region,omitempty"`
	ShippingCity                         string                       `bson:"shipping_city" json:"shipping_city,omitempty"`
	ShippingPostalCode                   string                       `bson:"shipping_postal_code" json:"shipping_postal_code,omitempty"`
	ShippingAddressLine1                 string                       `bson:"shipping_address_line1" json:"shipping_address_line1,omitempty"`
	ShippingAddressLine2                 string                       `bson:"shipping_address_line2" json:"shipping_address_line2,omitempty"`
	ShippingPostOfficeBoxNumber          string                       `bson:"shipping_post_office_box_number" json:"shipping_post_office_box_number"`
	ShippingFullAddressWithoutPostalCode string                       `bson:"shipping_full_address_without_postal_code" json:"shipping_full_address_without_postal_code,omitempty"` // Compiled value
	ShippingFullAddressWithPostalCode    string                       `bson:"shipping_full_address_with_postal_code" json:"shipping_full_address_with_postal_code,omitempty"`       // Compiled value
	ShippingFullAddressURL               string                       `bson:"shipping_full_address_url" json:"shipping_full_address_url,omitempty"`                                 // Compiled value
	HowDidYouHearAboutUsID               primitive.ObjectID           `bson:"how_did_you_hear_about_us_id" json:"how_did_you_hear_about_us_id,omitempty"`
	HowDidYouHearAboutUsText             string                       `bson:"how_did_you_hear_about_us_text" json:"how_did_you_hear_about_us_text,omitempty"`
	IsHowDidYouHearAboutUsOther          bool                         `bson:"is_how_did_you_hear_about_us_other" json:"is_how_did_you_hear_about_us_other,omitempty"`
	HowDidYouHearAboutUsOther            string                       `bson:"how_did_you_hear_about_us_other" json:"how_did_you_hear_about_us_other,omitempty"`
	AgreeTOS                             bool                         `bson:"agree_tos" json:"agree_tos,omitempty"`
	CreatedAt                            time.Time                    `bson:"created_at" json:"created_at,omitempty"`
	CreatedByUserID                      primitive.ObjectID           `bson:"created_by_user_id" json:"created_by_user_id"`
	CreatedByUserName                    string                       `bson:"created_by_user_name" json:"created_by_user_name"`
	CreatedFromIPAddress                 string                       `bson:"created_from_ip_address" json:"created_from_ip_address"`
	ModifiedAt                           time.Time                    `bson:"modified_at" json:"modified_at,omitempty"`
	ModifiedByUserID                     primitive.ObjectID           `bson:"modified_by_user_id" json:"modified_by_user_id"`
	ModifiedByUserName                   string                       `bson:"modified_by_user_name" json:"modified_by_user_name"`
	ModifiedFromIPAddress                string                       `bson:"modified_from_ip_address" json:"modified_from_ip_address"`
	Status                               int8                         `bson:"status" json:"status"`
	Salt                                 string                       `bson:"salt" json:"salt,omitempty"`
	JoinedTime                           time.Time                    `bson:"joined_time" json:"joined_time,omitempty"`
	PrAccessCode                         string                       `bson:"pr_access_code" json:"pr_access_code,omitempty"`
	PrExpiryTime                         time.Time                    `bson:"pr_expiry_time" json:"pr_expiry_time,omitempty"`
	RoleID                               int8                         `bson:"role_id" json:"role_id,omitempty"`
	Timezone                             string                       `bson:"timezone" json:"timezone,omitempty"`
	HasUserAccount                       bool                         `bson:"has_user_account" json:"has_user_account,omitempty"`
	UserID                               primitive.ObjectID           `bson:"user_id" json:"user_id,omitempty"`
	Type                                 int8                         `bson:"type" json:"type"`
	IsOkToEmail                          bool                         `bson:"is_ok_to_email" json:"is_ok_to_email"`
	IsOkToText                           bool                         `bson:"is_ok_to_text" json:"is_ok_to_text"`
	IsBusiness                           bool                         `bson:"is_business" json:"is_business"`
	IsSenior                             bool                         `bson:"is_senior" json:"is_senior"`
	IsSupport                            bool                         `bson:"is_support" json:"is_support"`
	JobInfoRead                          string                       `bson:"job_info_read" json:"job_info_read"`
	DeactivationReason                   int8                         `bson:"deactivation_reason" json:"deactivation_reason"`
	DeactivationReasonOther              string                       `bson:"deactivation_reason_other" json:"deactivation_reason_other"`
	Description                          string                       `bson:"description" json:"description"`
	AvatarObjectKey                      string                       `bson:"avatar_object_key" json:"avatar_object_key"`
	AvatarFileType                       string                       `bson:"avatar_file_type" json:"avatar_file_type"`
	AvatarFileName                       string                       `bson:"avatar_file_name" json:"avatar_file_name"`
	Birthdate                            time.Time                    `bson:"birthdate" json:"birthdate"`
	JoinDate                             time.Time                    `bson:"join_date" json:"join_date"`
	Nationality                          string                       `bson:"nationality" json:"nationality"`
	Gender                               string                       `bson:"gender" json:"gender"`
	TaxID                                string                       `bson:"tax_id" json:"tax_id"`
	Elevation                            float64                      `bson:"elevation" json:"elevation"`
	Latitude                             float64                      `bson:"latitude" json:"latitude"`
	Longitude                            float64                      `bson:"longitude" json:"longitude"`
	AreaServed                           string                       `bson:"area_served" json:"area_served"`
	AvailableLanguage                    string                       `bson:"available_language" json:"available_language"`
	ContactType                          string                       `bson:"contact_type" json:"contact_type"`
	OldID                                uint64                       `bson:"old_id" json:"old_id,omitempty"`
	HourlySalaryDesired                  int                          `bson:"hourly_salary_desired" json:"hourly_salary_desired"`
	LimitSpecial                         string                       `bson:"limit_special" json:"limit_special"`
	DuesDate                             time.Time                    `bson:"dues_date" json:"dues_date"`
	CommercialInsuranceExpiryDate        time.Time                    `bson:"commercial_insurance_expiry_date" json:"commercial_insurance_expiry_date"`
	AutoInsuranceExpiryDate              time.Time                    `bson:"auto_insurance_expiry_date" json:"auto_insurance_expiry_date"`
	WsibNumber                           string                       `bson:"wsib_number" json:"wsib_number"`
	WsibInsuranceDate                    time.Time                    `bson:"wsib_insurance_date" json:"wsib_insurance_date"`
	PoliceCheck                          time.Time                    `bson:"police_check" json:"police_check"`
	DriversLicenseClass                  string                       `bson:"drivers_license_class" json:"drivers_license_class"`
	Score                                float64                      `bson:"score" json:"score"`
	ServiceFeeID                         uint64                       `bson:"service_fee_id" json:"service_fee_id"`
	BalanceOwingAmount                   float64                      `bson:"balance_owing_amount" json:"balance_owing_amount"`
	EmergencyContactName                 string                       `bson:"emergency_contact_name" json:"emergency_contact_name"`
	EmergencyContactRelationship         string                       `bson:"emergency_contact_relationship" json:"emergency_contact_relationship"`
	EmergencyContactTelephone            string                       `bson:"emergency_contact_telephone" json:"emergency_contact_telephone"`
	EmergencyContactAlternativeTelephone string                       `bson:"emergency_contact_alternative_telephone" json:"emergency_contact_alternative_telephone"`
	Comments                             []*StaffComment              `bson:"comments" json:"comments"`
	SkillSets                            []*StaffSkillSet             `bson:"skill_sets" json:"skill_sets,omitempty"`
	InsuranceRequirements                []*StaffInsuranceRequirement `bson:"insurance_requirements" json:"insurance_requirements,omitempty"`
	VehicleTypes                         []*StaffVehicleType          `bson:"vehicle_types" json:"vehicle_types,omitempty"`
	AwayLogs                             []*StaffAwayLog              `bson:"away_logs" json:"away_logs,omitempty"`
	Tags                                 []*StaffTag                  `bson:"tags" json:"tags,omitempty"`
	// ServiceFee            *WorkOrderServiceFee             `json:"invoice_service_fee,omitempty"` // Referenced value from 'work_order_service_fee'.
	// Tags                  []*StaffTag                  `json:"tags,omitempty"`
}

type StaffComment struct {
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

type StaffVehicleType struct {
	ID          primitive.ObjectID `bson:"_id" json:"id"`
	TenantID    primitive.ObjectID `bson:"tenant_id" json:"tenant_id,omitempty"`
	Text        string             `bson:"text" json:"text"`
	Description string             `bson:"description" json:"description"`
	Status      int8               `bson:"status" json:"status"`
	OldID       uint64             `bson:"old_id" json:"old_id"`
}

type StaffSkillSet struct {
	ID          primitive.ObjectID `bson:"_id" json:"id"`
	TenantID    primitive.ObjectID `bson:"tenant_id" json:"tenant_id,omitempty"`
	Category    string             `bson:"category" json:"category"`
	SubCategory string             `bson:"sub_category" json:"sub_category"`
	Description string             `bson:"description" json:"description"`
	Status      int8               `bson:"status" json:"status"`
	OldID       uint64             `bson:"old_id" json:"old_id"`
}

type StaffInsuranceRequirement struct {
	ID          primitive.ObjectID `bson:"_id" json:"id"`
	TenantID    primitive.ObjectID `bson:"tenant_id" json:"tenant_id,omitempty"`
	Text        string             `bson:"text" json:"text"`
	Description string             `bson:"description" json:"description"`
	Status      int8               `bson:"status" json:"status"`
	OldID       uint64             `bson:"old_id" json:"old_id"`
}

type StaffAwayLog struct {
	ID                    primitive.ObjectID `bson:"_id" json:"id"`
	TenantID              primitive.ObjectID `bson:"tenant_id" json:"tenant_id,omitempty"`
	StaffID               primitive.ObjectID `bson:"associate_id" json:"associate_id"`
	StaffName             string             `bson:"associate_name" json:"associate_name,omitempty"`
	StaffLexicalName      string             `bson:"associate_lexical_name" json:"associate_lexical_name,omitempty"`
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

type StaffTag struct {
	ID          primitive.ObjectID `bson:"_id" json:"id"`
	TenantID    primitive.ObjectID `bson:"tenant_id" json:"tenant_id,omitempty"`
	Text        string             `bson:"text" json:"text"`
	Description string             `bson:"description" json:"description"`
	Status      int8               `bson:"status" json:"status"`
	OldID       uint64             `bson:"old_id" json:"old_id"`
}

type StaffListFilter struct {
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

type StaffListResult struct {
	Results     []*Staff           `json:"results"`
	NextCursor  primitive.ObjectID `json:"next_cursor"`
	HasNextPage bool               `json:"has_next_page"`
}

type StaffAsSelectOption struct {
	Value primitive.ObjectID `bson:"_id" json:"value"` // Extract from the database `_id` field and output through API as `value`.
	Label string             `bson:"name" json:"label"`
}

// StaffStorer Interface for user.
type StaffStorer interface {
	Create(ctx context.Context, m *Staff) error
	GetByID(ctx context.Context, id primitive.ObjectID) (*Staff, error)
	GetByOldID(ctx context.Context, oldID uint64) (*Staff, error)
	GetByEmail(ctx context.Context, email string) (*Staff, error)
	GetByVerificationCode(ctx context.Context, verificationCode string) (*Staff, error)
	CheckIfExistsByEmail(ctx context.Context, email string) (bool, error)
	UpdateByID(ctx context.Context, m *Staff) error
	UpsertByID(ctx context.Context, user *Staff) error
	ListByFilter(ctx context.Context, f *StaffListFilter) (*StaffListResult, error)
	ListAsSelectOptionByFilter(ctx context.Context, f *StaffListFilter) ([]*StaffAsSelectOption, error)
	DeleteByID(ctx context.Context, id primitive.ObjectID) error
	// //TODO: Add more...
}

type StaffStorerImpl struct {
	Logger     *slog.Logger
	DbClient   *mongo.Client
	Collection *mongo.Collection
}

func NewDatastore(appCfg *c.Conf, loggerp *slog.Logger, client *mongo.Client) StaffStorer {
	// ctx := context.Background()
	uc := client.Database(appCfg.DB.Name).Collection("staff")

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

	s := &StaffStorerImpl{
		Logger:     loggerp,
		DbClient:   client,
		Collection: uc,
	}
	return s
}

var StaffStateLabels = map[int8]string{
	StaffStatusActive:   "Active",
	StaffStatusArchived: "Archived",
}

var StaffTypeLabels = map[int8]string{
	StaffTypeResidential: "Residential",
	StaffTypeCommercial:  "Commercial",
	StaffTypeUnassigned:  "Unassigned",
}

var StaffDeactivationReasonLabels = map[int8]string{
	StaffDeactivationReasonNotSpecified:  "Not Specified",
	StaffDeactivationReasonOther:         "Other",
	StaffDeactivationReasonBlacklisted:   "Blacklisted",
	StaffDeactivationReasonMoved:         "Moved",
	StaffDeactivationReasonDeceased:      "Deceased",
	StaffDeactivationReasonDoNotConstact: "Do not contact",
}

var StaffTelephoneTypeLabels = map[int8]string{
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
