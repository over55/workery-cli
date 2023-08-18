package datastore

import (
	"context"
	"log"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/exp/slog"

	c "github.com/over55/workery-cli/config"
)

const (
	ServiceFeeStatusActive   = 1
	ServiceFeeStatusArchived = 100
)

type ServiceFee struct {
	ID          primitive.ObjectID `bson:"_id" json:"id"`
	TenantID    primitive.ObjectID `bson:"tenant_id" json:"tenant_id,omitempty"`
	Text        string             `bson:"text" json:"text"`
	Description string             `bson:"description" json:"description"`
	Percentage  float64            `bson:"percentage" json:"percentage"`
	Status      int8               `bson:"status" json:"status"`
	OldID       uint64             `bson:"old_id" json:"old_id"`
}

type ServiceFeeListFilter struct {
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

type ServiceFeeListResult struct {
	Results     []*ServiceFee      `json:"results"`
	NextCursor  primitive.ObjectID `json:"next_cursor"`
	HasNextPage bool               `json:"has_next_page"`
}

type ServiceFeeAsSelectOption struct {
	Value primitive.ObjectID `bson:"_id" json:"value"` // Extract from the database `_id` field and output through API as `value`.
	Label string             `bson:"name" json:"label"`
}

// ServiceFeeStorer Interface for user.
type ServiceFeeStorer interface {
	Create(ctx context.Context, m *ServiceFee) error
	GetByID(ctx context.Context, id primitive.ObjectID) (*ServiceFee, error)
	GetByOldID(ctx context.Context, oldID uint64) (*ServiceFee, error)
	GetByEmail(ctx context.Context, email string) (*ServiceFee, error)
	GetByVerificationCode(ctx context.Context, verificationCode string) (*ServiceFee, error)
	CheckIfExistsByEmail(ctx context.Context, email string) (bool, error)
	UpdateByID(ctx context.Context, m *ServiceFee) error
	ListByFilter(ctx context.Context, f *ServiceFeeListFilter) (*ServiceFeeListResult, error)
	ListAsSelectOptionByFilter(ctx context.Context, f *ServiceFeeListFilter) ([]*ServiceFeeAsSelectOption, error)
	DeleteByID(ctx context.Context, id primitive.ObjectID) error
	// //TODO: Add more...
}

type ServiceFeeStorerImpl struct {
	Logger     *slog.Logger
	DbClient   *mongo.Client
	Collection *mongo.Collection
}

func NewDatastore(appCfg *c.Conf, loggerp *slog.Logger, client *mongo.Client) ServiceFeeStorer {
	// ctx := context.Background()
	uc := client.Database(appCfg.DB.Name).Collection("service_fees")

	// The following few lines of code will create the index for our app for this
	// colleciton.
	indexModel := mongo.IndexModel{
		Keys: bson.D{
			{"text", "text"},
			{"description", "text"},
		},
	}
	_, err := uc.Indexes().CreateOne(context.TODO(), indexModel)
	if err != nil {
		// It is important that we crash the app on startup to meet the
		// requirements of `google/wire` framework.
		log.Fatal(err)
	}

	s := &ServiceFeeStorerImpl{
		Logger:     loggerp,
		DbClient:   client,
		Collection: uc,
	}
	return s
}
