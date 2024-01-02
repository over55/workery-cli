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

const (
	ActivitySheetStatusPending  = 5
	ActivitySheetStatusDeclined = 4
	ActivitySheetStatusAccepted = 3
	ActivitySheetStatusError    = 2
	ActivitySheetStatusArchived = 1
	ActivitySheetTypeUnassigned = 1
	ActivitySheetTypeResidentia = 2
	ActivitySheetTypeCommercial = 3
)

type ActivitySheet struct {
	ID                    primitive.ObjectID `bson:"_id" json:"id"`
	OrderID               primitive.ObjectID `bson:"order_id" json:"order_id"`
	OrderWJID             uint64             `bson:"order_wjid" json:"order_wjid"` // A.K.A. `Workery Job ID`
	AssociateID           primitive.ObjectID `bson:"associate_id" json:"associate_id"`
	AssociateName         string             `bson:"associate_name" json:"associate_name"`
	AssociateLexicalName  string             `bson:"associate_lexical_name" json:"associate_lexical_name"`
	Comment               string             `bson:"comment" json:"comment"`
	CreatedAt             time.Time          `bson:"created_at" json:"created_at"`
	CreatedByUserID       primitive.ObjectID `bson:"created_by_user_id" json:"created_by_user_id,omitempty"`
	CreatedByUserName     string             `bson:"created_by_user_name" json:"created_by_user_name"`
	CreatedFromIPAddress  string             `bson:"created_from_ip_address" json:"created_from_ip_address"`
	ModifiedAt            time.Time          `bson:"modified_at" json:"modified_at"`
	ModifiedByUserID      primitive.ObjectID `bson:"modified_by_user_id" json:"modified_by_user_id,omitempty"`
	ModifiedByUserName    string             `bson:"modified_by_user_name" json:"modified_by_user_name"`
	ModifiedFromIPAddress string             `bson:"modified_from_ip_address" json:"modified_from_ip_address"`
	Status                int8               `bson:"status" json:"status"`
	Type                  int8               `bson:"type_of" json:"type_of"`
	TenantID              primitive.ObjectID `bson:"tenant_id" json:"tenant_id,omitempty"`
	OrderTenantIDWithWJID string             `bson:"order_tenant_id_with_wjid" json:"order_tenant_id_with_wjid"` // OrderTenantIDWithWJID is a combination of `tenancy_id` and `wjid` values written in the following structure `%v_%v`.
	PublicID                 uint64             `bson:"public_id" json:"public_id"`
	// OngoingOrderID        primitive.ObjectID `bson:"ongoing_order_id" json:"ongoing_order_id"`
}

type ActivitySheetListFilter struct {
	// Pagination related.
	Cursor    primitive.ObjectID
	PageSize  int64
	SortField string
	SortOrder int8 // 1=ascending | -1=descending

	// Filter related.
	TenantID        primitive.ObjectID
	Status          int8
	ExcludeArchived bool
	SearchText      string
}

type ActivitySheetListResult struct {
	Results     []*ActivitySheet   `json:"results"`
	NextCursor  primitive.ObjectID `json:"next_cursor"`
	HasNextPage bool               `json:"has_next_page"`
}

type ActivitySheetAsSelectOption struct {
	Value primitive.ObjectID `bson:"_id" json:"value"` // Extract from the database `_id` field and output through API as `value`.
	Label string             `bson:"name" json:"label"`
}

// ActivitySheetStorer Interface for user.
type ActivitySheetStorer interface {
	Create(ctx context.Context, m *ActivitySheet) error
	GetByID(ctx context.Context, id primitive.ObjectID) (*ActivitySheet, error)
	GetByPublicID(ctx context.Context, oldID uint64) (*ActivitySheet, error)
	GetByEmail(ctx context.Context, email string) (*ActivitySheet, error)
	GetByVerificationCode(ctx context.Context, verificationCode string) (*ActivitySheet, error)
	CheckIfExistsByEmail(ctx context.Context, email string) (bool, error)
	UpdateByID(ctx context.Context, m *ActivitySheet) error
	ListByFilter(ctx context.Context, f *ActivitySheetListFilter) (*ActivitySheetListResult, error)
	ListAsSelectOptionByFilter(ctx context.Context, f *ActivitySheetListFilter) ([]*ActivitySheetAsSelectOption, error)
	DeleteByID(ctx context.Context, id primitive.ObjectID) error
	// //TODO: Add more...
}

type ActivitySheetStorerImpl struct {
	Logger     *slog.Logger
	DbClient   *mongo.Client
	Collection *mongo.Collection
}

func NewDatastore(appCfg *c.Conf, loggerp *slog.Logger, client *mongo.Client) ActivitySheetStorer {
	// ctx := context.Background()
	uc := client.Database(appCfg.DB.Name).Collection("activity_sheets")

	// The following few lines of code will create the index for our app for this
	// colleciton.
	indexModel := mongo.IndexModel{
		Keys: bson.D{
			{"comment", "text"},
			{"associate_name", "text"},
		},
	}
	_, err := uc.Indexes().CreateOne(context.TODO(), indexModel)
	if err != nil {
		// It is important that we crash the app on startup to meet the
		// requirements of `google/wire` framework.
		log.Fatal(err)
	}

	s := &ActivitySheetStorerImpl{
		Logger:     loggerp,
		DbClient:   client,
		Collection: uc,
	}
	return s
}
