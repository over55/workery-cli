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
	CustomerStatusActive   = 1
	CustomerStatusArchived = 2

	CustomerDeactivationReasonNotSpecified  = 0
	CustomerDeactivationReasonOther         = 1
	CustomerDeactivationReasonBlacklisted   = 2
	CustomerDeactivationReasonMoved         = 3
	CustomerDeactivationReasonDeceased      = 4
	CustomerDeactivationReasonDoNotConstact = 5

	CustomerTypeUnassigned  = 1
	CustomerTypeResidential = 2
	CustomerTypeCommercial  = 3

	CustomerPhoneTypeLandline = 1
	CustomerPhoneTypeMobile   = 2
	CustomerPhoneTypeWork     = 3
)

type Customer struct {
	ID                                   primitive.ObjectID `bson:"_id" json:"id"`
	TenantID                             primitive.ObjectID `bson:"tenant_id" json:"tenant_id,omitempty"`
	FirstName                            string             `bson:"first_name" json:"first_name"`
	LastName                             string             `bson:"last_name" json:"last_name"`
	Name                                 string             `bson:"name" json:"name"`
	LexicalName                          string             `bson:"lexical_name" json:"lexical_name"`
	Email                                string             `bson:"email" json:"email"`
	Phone                                string             `bson:"phone" json:"phone,omitempty"`
	PhoneType                            int8               `bson:"phone_type" json:"phone_type"`
	PhoneExtension                       string             `bson:"phone_extension" json:"phone_extension"`
	FaxNumber                            string             `bson:"fax_number" json:"fax_number"`
	OtherPhone                           string             `bson:"other_phone" json:"other_phone"`
	OtherPhoneExtension                  string             `bson:"other_phone_extension" json:"other_phone_extension"`
	OtherPhoneType                       int8               `bson:"other_phone_type" json:"other_phone_type"`
	Country                              string             `bson:"country" json:"country,omitempty"`
	Region                               string             `bson:"region" json:"region,omitempty"`
	City                                 string             `bson:"city" json:"city,omitempty"`
	PostalCode                           string             `bson:"postal_code" json:"postal_code,omitempty"`
	AddressLine1                         string             `bson:"address_line1" json:"address_line1,omitempty"`
	AddressLine2                         string             `bson:"address_line2" json:"address_line2,omitempty"`
	PostOfficeBoxNumber                  string             `bson:"post_office_box_number" json:"post_office_box_number"`
	FullAddressWithoutPostalCode         string             `bson:"full_address_without_postal_code" json:"full_address_without_postal_code,omitempty"` // Compiled value
	FullAddressWithPostalCode            string             `bson:"full_address_with_postal_code" json:"full_address_with_postal_code,omitempty"`       // Compiled value
	FullAddressURL                       string             `bson:"full_address_url" json:"full_address_url,omitempty"`                                 // Compiled value
	HasShippingAddress                   bool               `bson:"has_shipping_address" json:"has_shipping_address,omitempty"`
	ShippingName                         string             `bson:"shipping_name" json:"shipping_name,omitempty"`
	ShippingPhone                        string             `bson:"shipping_phone" json:"shipping_phone,omitempty"`
	ShippingCountry                      string             `bson:"shipping_country" json:"shipping_country,omitempty"`
	ShippingRegion                       string             `bson:"shipping_region" json:"shipping_region,omitempty"`
	ShippingCity                         string             `bson:"shipping_city" json:"shipping_city,omitempty"`
	ShippingPostalCode                   string             `bson:"shipping_postal_code" json:"shipping_postal_code,omitempty"`
	ShippingAddressLine1                 string             `bson:"shipping_address_line1" json:"shipping_address_line1,omitempty"`
	ShippingAddressLine2                 string             `bson:"shipping_address_line2" json:"shipping_address_line2,omitempty"`
	ShippingPostOfficeBoxNumber          string             `bson:"shipping_post_office_box_number" json:"shipping_post_office_box_number"`
	ShippingFullAddressWithoutPostalCode string             `bson:"shipping_full_address_without_postal_code" json:"shipping_full_address_without_postal_code,omitempty"` // Compiled value
	ShippingFullAddressWithPostalCode    string             `bson:"shipping_full_address_with_postal_code" json:"shipping_full_address_with_postal_code,omitempty"`       // Compiled value
	ShippingFullAddressURL               string             `bson:"shipping_full_address_url" json:"shipping_full_address_url,omitempty"`                                 // Compiled value
	HowDidYouHearAboutUsID               primitive.ObjectID `bson:"how_did_you_hear_about_us_id" json:"how_did_you_hear_about_us_id,omitempty"`
	HowDidYouHearAboutUsOther            string             `bson:"how_did_you_hear_about_us_other" json:"how_did_you_hear_about_us_other,omitempty"`
	HowDidYouHearAboutUsValue            string             `bson:"how_did_you_hear_about_us_value" json:"how_did_you_hear_about_us_value,omitempty"`
	AgreeTOS                             bool               `bson:"agree_tos" json:"agree_tos,omitempty"`
	CreatedAt                            time.Time          `bson:"created_at" json:"created_at,omitempty"`
	CreatedByUserID                      primitive.ObjectID `bson:"created_by_user_id" json:"created_by_user_id"`
	CreatedByUserName                    string             `bson:"created_by_user_name" json:"created_by_user_name"`
	CreatedFromIPAddress                 string             `bson:"created_from_ip_address" json:"created_from_ip_address"`
	ModifiedAt                           time.Time          `bson:"modified_at" json:"modified_at,omitempty"`
	ModifiedByUserID                     primitive.ObjectID `bson:"modified_by_user_id" json:"modified_by_user_id"`
	ModifiedByUserName                   string             `bson:"modified_by_user_name" json:"modified_by_user_name"`
	ModifiedFromIPAddress                string             `bson:"modified_from_ip_address" json:"modified_from_ip_address"`
	Status                               int8               `bson:"status" json:"status"`
	Salt                                 string             `bson:"salt" json:"salt,omitempty"`
	JoinedTime                           time.Time          `bson:"joined_time" json:"joined_time,omitempty"`
	PrAccessCode                         string             `bson:"pr_access_code" json:"pr_access_code,omitempty"`
	PrExpiryTime                         time.Time          `bson:"pr_expiry_time" json:"pr_expiry_time,omitempty"`
	RoleID                               int8               `bson:"role_id" json:"role_id,omitempty"`
	Timezone                             string             `bson:"timezone" json:"timezone,omitempty"`
	HasUserAccount                       bool               `bson:"has_user_account" json:"has_user_account,omitempty"`
	UserID                               primitive.ObjectID `bson:"user_id" json:"user_id,omitempty"`
	Type                                 int8               `bson:"type" json:"type"`
	IsOkToEmail                          bool               `bson:"is_ok_to_email" json:"is_ok_to_email"`
	IsOkToText                           bool               `bson:"is_ok_to_text" json:"is_ok_to_text"`
	IsBusiness                           bool               `bson:"is_business" json:"is_business"`
	IsSenior                             bool               `bson:"is_senior" json:"is_senior"`
	IsSupport                            bool               `bson:"is_support" json:"is_support"`
	JobInfoRead                          string             `bson:"job_info_read" json:"job_info_read"`
	DeactivationReason                   int8               `bson:"deactivation_reason" json:"deactivation_reason"`
	DeactivationReasonOther              string             `bson:"deactivation_reason_other" json:"deactivation_reason_other"`
	Description                          string             `bson:"description" json:"description"`
	AvatarObjectKey                      string             `bson:"avatar_object_key" json:"avatar_object_key"`
	AvatarFileType                       string             `bson:"avatar_file_type" json:"avatar_file_type"`
	AvatarFileName                       string             `bson:"avatar_file_name" json:"avatar_file_name"`
	Birthdate                            time.Time          `bson:"birthdate" json:"birthdate"`
	JoinDate                             time.Time          `bson:"join_date" json:"join_date"`
	Nationality                          string             `bson:"nationality" json:"nationality"`
	Gender                               string             `bson:"gender" json:"gender"`
	TaxId                                string             `bson:"tax_id" json:"tax_id"`
	Elevation                            float64            `bson:"elevation" json:"elevation"`
	Latitude                             float64            `bson:"latitude" json:"latitude"`
	Longitude                            float64            `bson:"longitude" json:"longitude"`
	AreaServed                           string             `bson:"area_served" json:"area_served"`
	AvailableLanguage                    string             `bson:"available_language" json:"available_language"`
	ContactType                          string             `bson:"contact_type" json:"contact_type"`
	OrganizationName                     string             `bson:"organization_name" json:"organization_name"`
	OrganizationType                     int8               `bson:"organization_type" json:"organization_type"`
	OldID                                uint64             `bson:"old_id" json:"old_id,omitempty"`
	Comments                             []*CustomerComment `bson:"comments" json:"comments"`
	Tags                                 []*CustomerTag     `bson:"tags" json:"tags"`
	//TODO: Add references here...
}

type CustomerComment struct {
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
	Text                  string             `bson:"text" json:"text"`
	Status                int8               `bson:"status" json:"status"`
	OldID                 uint64             `bson:"old_id" json:"old_id"`
}

type CustomerTag struct {
	ID          primitive.ObjectID `bson:"_id" json:"id"`
	TenantID    primitive.ObjectID `bson:"tenant_id" json:"tenant_id,omitempty"`
	Text        string             `bson:"text" json:"text"`
	Description string             `bson:"description" json:"description"`
	Status      int8               `bson:"status" json:"status"`
	OldID       uint64             `bson:"old_id" json:"old_id"`
}

type CustomerListFilter struct {
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

type CustomerListResult struct {
	Results     []*Customer        `json:"results"`
	NextCursor  primitive.ObjectID `json:"next_cursor"`
	HasNextPage bool               `json:"has_next_page"`
}

type CustomerAsSelectOption struct {
	Value primitive.ObjectID `bson:"_id" json:"value"` // Extract from the database `_id` field and output through API as `value`.
	Label string             `bson:"name" json:"label"`
}

// CustomerStorer Interface for user.
type CustomerStorer interface {
	Create(ctx context.Context, m *Customer) error
	GetByID(ctx context.Context, id primitive.ObjectID) (*Customer, error)
	GetByOldID(ctx context.Context, oldID uint64) (*Customer, error)
	GetByEmail(ctx context.Context, email string) (*Customer, error)
	GetByVerificationCode(ctx context.Context, verificationCode string) (*Customer, error)
	CheckIfExistsByEmail(ctx context.Context, email string) (bool, error)
	UpdateByID(ctx context.Context, m *Customer) error
	UpsertByID(ctx context.Context, user *Customer) error
	ListByFilter(ctx context.Context, f *CustomerListFilter) (*CustomerListResult, error)
	ListAsSelectOptionByFilter(ctx context.Context, f *CustomerListFilter) ([]*CustomerAsSelectOption, error)
	DeleteByID(ctx context.Context, id primitive.ObjectID) error
	// //TODO: Add more...
}

type CustomerStorerImpl struct {
	Logger     *slog.Logger
	DbClient   *mongo.Client
	Collection *mongo.Collection
}

func NewDatastore(appCfg *c.Conf, loggerp *slog.Logger, client *mongo.Client) CustomerStorer {
	// ctx := context.Background()
	uc := client.Database(appCfg.DB.Name).Collection("customers")

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

	s := &CustomerStorerImpl{
		Logger:     loggerp,
		DbClient:   client,
		Collection: uc,
	}
	return s
}

var CustomerStateLabels = map[int8]string{
	CustomerStatusActive:   "Active",
	CustomerStatusArchived: "Archived",
}

var CustomerTypeOfLabels = map[int8]string{
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

var CustomerTelephoneTypeOfLabels = map[int8]string{
	1: "Landline",
	2: "Mobile",
	3: "Work",
}

//---------------------
// organization_type_of
//---------------------
// 1 = Unknown Organization Type | UNKNOWN_ORGANIZATION_TYPE_OF_ID
// 2 = Private Organization Type | PRIVATE_ORGANIZATION_TYPE_OF_ID
// 3 = Non-Profit Organization Type | NON_PROFIT_ORGANIZATION_TYPE_OF_ID
// 4 = Government Organization | GOVERNMENT_ORGANIZATION_TYPE_OF_ID
