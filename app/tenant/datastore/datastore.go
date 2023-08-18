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
	TenantPendingStatus  = 1
	TenantActiveStatus   = 2
	TenantErrorStatus    = 3
	TenantArchivedStatus = 4
	RootType             = 1
	RetailerType         = 2
)

type Tenant struct {
	ID                      primitive.ObjectID `bson:"_id" json:"id"`
	Uuid                    string             `bson:"uuid" json:"uuid"`
	SchemaName              string             `bson:"schema_name" json:"schema_name"`
	AlternateName           string             `bson:"alternate_name" json:"alternate_name"`
	Description             string             `bson:"description" json:"description"`
	Name                    string             `bson:"name" json:"name"`
	Url                     string             `bson:"url" json:"url"`
	Status                  int8               `bson:"status" json:"status"`
	Timezone                string             `bson:"timestamp" json:"timestamp"`
	CreatedTime             time.Time          `bson:"created_time" json:"created_time"`
	ModifiedTime            time.Time          `bson:"modified_time" json:"modified_time"`
	AddressCountry          string             `bson:"address_country" json:"address_country"`
	AddressRegion           string             `bson:"address_region" json:"address_region"`
	AddressLocality         string             `bson:"address_locality" json:"address_locality"`
	PostOfficeBoxNumber     string             `bson:"post_office_box_number" json:"post_office_box_number"`
	PostalCode              string             `bson:"postal_code" json:"postal_code"`
	StreetAddress           string             `bson:"street_address" json:"street_address"`
	StreetAddressExtra      string             `bson:"street_address_extra" json:"street_address_extra"`
	Elevation               float64            `bson:"elevation" json:"elevation"`
	Latitude                float64            `bson:"latitude" json:"latitude"`
	Longitude               float64            `bson:"longitude" json:"longitude"`
	AreaServed              string             `bson:"area_served" json:"area_served"`
	AvailableLanguage       string             `bson:"available_language" json:"available_language"`
	ContactType             string             `bson:"contact_type" json:"contact_type"`
	Email                   string             `bson:"email" json:"email"`
	FaxNumber               string             `bson:"fax_number" json:"fax_number"`
	Telephone               string             `bson:"telephone" json:"telephone"`
	TelephoneTypeOf         int8               `bson:"telephone_type_of" json:"telephone_type_of"`
	TelephoneExtension      string             `bson:"telephone_extension" json:"telephone_extension"`
	OtherTelephone          string             `bson:"other_telephone" json:"other_telephone"`
	OtherTelephoneExtension string             `bson:"other_telephone_extension" json:"other_telephone_extension"`
	OtherTelephoneTypeOf    int8               `bson:"other_telephone_type_of" json:"other_telephone_type_of"`
	OldId                   uint64             `bson:"old_id" json:"old_id"`
}

type TenantListFilter struct {
	// Pagination related.
	Cursor    primitive.ObjectID
	PageSize  int64
	SortField string
	SortOrder int8 // 1=ascending | -1=descending

	// Filter related.
	TenantID        primitive.ObjectID
	UserID          primitive.ObjectID
	UserRole        int8
	Status          int8
	ExcludeArchived bool
	SearchText      string
	CreatedAtGTE    time.Time
}

type TenantListResult struct {
	Results     []*Tenant          `json:"results"`
	NextCursor  primitive.ObjectID `json:"next_cursor"`
	HasNextPage bool               `json:"has_next_page"`
}

type TenantAsSelectOption struct {
	Value primitive.ObjectID `bson:"_id" json:"value"` // Extract from the database `_id` field and output through API as `value`.
	Label string             `bson:"name" json:"label"`
}

// TenantStorer Interface for tenant.
type TenantStorer interface {
	Create(ctx context.Context, m *Tenant) error
	GetByID(ctx context.Context, id primitive.ObjectID) (*Tenant, error)
	GetByOldID(ctx context.Context, oldID uint64) (*Tenant, error)
	GetBySchemaName(ctx context.Context, schemaName string) (*Tenant, error)
	UpdateByID(ctx context.Context, m *Tenant) error
	ListByFilter(ctx context.Context, m *TenantListFilter) (*TenantListResult, error)
	ListAsSelectOptionByFilter(ctx context.Context, f *TenantListFilter) ([]*TenantAsSelectOption, error)
	DeleteByID(ctx context.Context, id primitive.ObjectID) error
	// //TODO: Add more...
}

type TenantStorerImpl struct {
	Logger     *slog.Logger
	DbClient   *mongo.Client
	Collection *mongo.Collection
}

func NewDatastore(appCfg *c.Conf, loggerp *slog.Logger, client *mongo.Client) TenantStorer {
	// ctx := context.Background()
	uc := client.Database(appCfg.DB.Name).Collection("tenants")

	// The following few lines of code will create the index for our app for this
	// colleciton.
	indexModel := mongo.IndexModel{
		Keys: bson.D{
			{"name", "text"},
		},
	}
	_, err := uc.Indexes().CreateOne(context.TODO(), indexModel)
	if err != nil {
		// It is important that we crash the app on startup to meet the
		// requirements of `google/wire` framework.
		log.Fatal(err)
	}

	s := &TenantStorerImpl{
		Logger:     loggerp,
		DbClient:   client,
		Collection: uc,
	}
	return s
}
