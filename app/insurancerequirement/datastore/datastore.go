package datastore

import (
	"context"
	"log"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"log/slog"

	c "github.com/over55/workery-cli/config"
)

const (
	InsuranceRequirementStatusActive   = 1
	InsuranceRequirementStatusArchived = 100
)

type InsuranceRequirement struct {
	ID          primitive.ObjectID `bson:"_id" json:"id"`
	TenantID    primitive.ObjectID `bson:"tenant_id" json:"tenant_id,omitempty"`
	Name        string             `bson:"name" json:"name"`
	Description string             `bson:"description" json:"description"`
	Status      int8               `bson:"status" json:"status"`
	OldID       uint64             `bson:"old_id" json:"old_id"`
}

type InsuranceRequirementListFilter struct {
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

type InsuranceRequirementListResult struct {
	Results     []*InsuranceRequirement `json:"results"`
	NextCursor  primitive.ObjectID      `json:"next_cursor"`
	HasNextPage bool                    `json:"has_next_page"`
}

type InsuranceRequirementAsSelectOption struct {
	Value primitive.ObjectID `bson:"_id" json:"value"` // Extract from the database `_id` field and output through API as `value`.
	Label string             `bson:"name" json:"label"`
}

// InsuranceRequirementStorer Interface for user.
type InsuranceRequirementStorer interface {
	Create(ctx context.Context, m *InsuranceRequirement) error
	GetByID(ctx context.Context, id primitive.ObjectID) (*InsuranceRequirement, error)
	GetByOldID(ctx context.Context, oldID uint64) (*InsuranceRequirement, error)
	GetByEmail(ctx context.Context, email string) (*InsuranceRequirement, error)
	GetByVerificationCode(ctx context.Context, verificationCode string) (*InsuranceRequirement, error)
	CheckIfExistsByEmail(ctx context.Context, email string) (bool, error)
	UpdateByID(ctx context.Context, m *InsuranceRequirement) error
	ListByFilter(ctx context.Context, f *InsuranceRequirementListFilter) (*InsuranceRequirementListResult, error)
	ListAsSelectOptionByFilter(ctx context.Context, f *InsuranceRequirementListFilter) ([]*InsuranceRequirementAsSelectOption, error)
	DeleteByID(ctx context.Context, id primitive.ObjectID) error
	// //TODO: Add more...
}

type InsuranceRequirementStorerImpl struct {
	Logger     *slog.Logger
	DbClient   *mongo.Client
	Collection *mongo.Collection
}

func NewDatastore(appCfg *c.Conf, loggerp *slog.Logger, client *mongo.Client) InsuranceRequirementStorer {
	// ctx := context.Background()
	uc := client.Database(appCfg.DB.Name).Collection("insurance_requirements")

	// The following few lines of code will create the index for our app for this
	// colleciton.
	indexModel := mongo.IndexModel{
		Keys: bson.D{
			{"name", "text"},
			{"description", "text"},
		},
	}
	_, err := uc.Indexes().CreateOne(context.TODO(), indexModel)
	if err != nil {
		// It is important that we crash the app on startup to meet the
		// requirements of `google/wire` framework.
		log.Fatal(err)
	}

	s := &InsuranceRequirementStorerImpl{
		Logger:     loggerp,
		DbClient:   client,
		Collection: uc,
	}
	return s
}
