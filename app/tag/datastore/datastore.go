package datastore

import (
	"context"
	"log"
	"log/slog"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	c "github.com/over55/workery-cli/config"
)

const (
	TagStatusActive   = 1
	TagStatusArchived = 2
)

type Tag struct {
	ID          primitive.ObjectID `bson:"_id" json:"id"`
	TenantID    primitive.ObjectID `bson:"tenant_id" json:"tenant_id,omitempty"`
	Text        string             `bson:"text" json:"text"`
	Description string             `bson:"description" json:"description"`
	Status      int8               `bson:"status" json:"status"`
	PublicID       uint64             `bson:"public_id" json:"public_id"`
}

type TagListFilter struct {
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

type TagListResult struct {
	Results     []*Tag             `json:"results"`
	NextCursor  primitive.ObjectID `json:"next_cursor"`
	HasNextPage bool               `json:"has_next_page"`
}

type TagAsSelectOption struct {
	Value primitive.ObjectID `bson:"_id" json:"value"` // Extract from the database `_id` field and output through API as `value`.
	Label string             `bson:"name" json:"label"`
}

// TagStorer Interface for user.
type TagStorer interface {
	Create(ctx context.Context, m *Tag) error
	GetByID(ctx context.Context, id primitive.ObjectID) (*Tag, error)
	GetByPublicID(ctx context.Context, oldID uint64) (*Tag, error)
	GetByEmail(ctx context.Context, email string) (*Tag, error)
	GetByVerificationCode(ctx context.Context, verificationCode string) (*Tag, error)
	CheckIfExistsByEmail(ctx context.Context, email string) (bool, error)
	UpdateByID(ctx context.Context, m *Tag) error
	ListByFilter(ctx context.Context, f *TagListFilter) (*TagListResult, error)
	ListAsSelectOptionByFilter(ctx context.Context, f *TagListFilter) ([]*TagAsSelectOption, error)
	DeleteByID(ctx context.Context, id primitive.ObjectID) error
	// //TODO: Add more...
}

type TagStorerImpl struct {
	Logger     *slog.Logger
	DbClient   *mongo.Client
	Collection *mongo.Collection
}

func NewDatastore(appCfg *c.Conf, loggerp *slog.Logger, client *mongo.Client) TagStorer {
	// ctx := context.Background()
	uc := client.Database(appCfg.DB.Name).Collection("tags")

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

	s := &TagStorerImpl{
		Logger:     loggerp,
		DbClient:   client,
		Collection: uc,
	}
	return s
}
