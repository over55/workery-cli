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
	VehicleTypeStatusActive   = 1
	VehicleTypeStatusArchived = 100
)

type VehicleType struct {
	ID          primitive.ObjectID `bson:"_id" json:"id"`
	TenantID    primitive.ObjectID `bson:"tenant_id" json:"tenant_id,omitempty"`
	Text        string             `bson:"text" json:"text"`
	Description string             `bson:"description" json:"description"`
	Status      int8               `bson:"status" json:"status"`
	OldID       uint64             `bson:"old_id" json:"old_id"`
}

type VehicleTypeListFilter struct {
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

type VehicleTypeListResult struct {
	Results     []*VehicleType     `json:"results"`
	NextCursor  primitive.ObjectID `json:"next_cursor"`
	HasNextPage bool               `json:"has_next_page"`
}

type VehicleTypeAsSelectOption struct {
	Value primitive.ObjectID `bson:"_id" json:"value"` // Extract from the database `_id` field and output through API as `value`.
	Label string             `bson:"name" json:"label"`
}

// VehicleTypeStorer Interface for user.
type VehicleTypeStorer interface {
	Create(ctx context.Context, m *VehicleType) error
	GetByID(ctx context.Context, id primitive.ObjectID) (*VehicleType, error)
	GetByOldID(ctx context.Context, oldID uint64) (*VehicleType, error)
	GetByEmail(ctx context.Context, email string) (*VehicleType, error)
	GetByVerificationCode(ctx context.Context, verificationCode string) (*VehicleType, error)
	CheckIfExistsByEmail(ctx context.Context, email string) (bool, error)
	UpdateByID(ctx context.Context, m *VehicleType) error
	ListByFilter(ctx context.Context, f *VehicleTypeListFilter) (*VehicleTypeListResult, error)
	ListAsSelectOptionByFilter(ctx context.Context, f *VehicleTypeListFilter) ([]*VehicleTypeAsSelectOption, error)
	DeleteByID(ctx context.Context, id primitive.ObjectID) error
	// //TODO: Add more...
}

type VehicleTypeStorerImpl struct {
	Logger     *slog.Logger
	DbClient   *mongo.Client
	Collection *mongo.Collection
}

func NewDatastore(appCfg *c.Conf, loggerp *slog.Logger, client *mongo.Client) VehicleTypeStorer {
	// ctx := context.Background()
	uc := client.Database(appCfg.DB.Name).Collection("vehicle_types")

	// The following few lines of code will create the index for our app for this
	// colleciton.
	indexModel := mongo.IndexModel{
		Keys: bson.D{
			{"tenant_name", "text"},
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

	s := &VehicleTypeStorerImpl{
		Logger:     loggerp,
		DbClient:   client,
		Collection: uc,
	}
	return s
}
