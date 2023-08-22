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
	HowHearAboutUsItemStatusActive   = 1
	HowHearAboutUsItemStatusArchived = 100
)

type HowHearAboutUsItem struct {
	ID             primitive.ObjectID `bson:"_id" json:"id"`
	TenantID       primitive.ObjectID `bson:"tenant_id" json:"tenant_id"`
	Text           string             `bson:"text" json:"text"`
	SortNumber     int8               `bson:"sort_number" json:"sort_number"`
	IsForAssociate bool               `bson:"is_for_associate" json:"is_for_associate"`
	IsForCustomer  bool               `bson:"is_for_customer" json:"is_for_customer"`
	IsForStaff     bool               `bson:"is_for_staff" json:"is_for_staff"`
	Status         int8               `bson:"status" json:"status"`
	OldID          uint64             `bson:"old_id" json:"old_id"`
}

type HowHearAboutUsItemListFilter struct {
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

type HowHearAboutUsItemListResult struct {
	Results     []*HowHearAboutUsItem `json:"results"`
	NextCursor  primitive.ObjectID    `json:"next_cursor"`
	HasNextPage bool                  `json:"has_next_page"`
}

type HowHearAboutUsItemAsSelectOption struct {
	Value primitive.ObjectID `bson:"_id" json:"value"` // Extract from the database `_id` field and output through API as `value`.
	Label string             `bson:"name" json:"label"`
}

// HowHearAboutUsItemStorer Interface for user.
type HowHearAboutUsItemStorer interface {
	Create(ctx context.Context, m *HowHearAboutUsItem) error
	GetByID(ctx context.Context, id primitive.ObjectID) (*HowHearAboutUsItem, error)
	GetByOldID(ctx context.Context, oldID uint64) (*HowHearAboutUsItem, error)
	GetByText(ctx context.Context, text string) (*HowHearAboutUsItem, error)
	CheckIfExistsByEmail(ctx context.Context, email string) (bool, error)
	UpdateByID(ctx context.Context, m *HowHearAboutUsItem) error
	ListByFilter(ctx context.Context, f *HowHearAboutUsItemListFilter) (*HowHearAboutUsItemListResult, error)
	ListAsSelectOptionByFilter(ctx context.Context, f *HowHearAboutUsItemListFilter) ([]*HowHearAboutUsItemAsSelectOption, error)
	DeleteByID(ctx context.Context, id primitive.ObjectID) error
	// //TODO: Add more...
}

type HowHearAboutUsItemStorerImpl struct {
	Logger     *slog.Logger
	DbClient   *mongo.Client
	Collection *mongo.Collection
}

func NewDatastore(appCfg *c.Conf, loggerp *slog.Logger, client *mongo.Client) HowHearAboutUsItemStorer {
	// ctx := context.Background()
	uc := client.Database(appCfg.DB.Name).Collection("how_hear_about_us_items")

	// The following few lines of code will create the index for our app for this
	// colleciton.
	indexModel := mongo.IndexModel{
		Keys: bson.D{
			{"text", "text"},
		},
	}
	_, err := uc.Indexes().CreateOne(context.TODO(), indexModel)
	if err != nil {
		// It is important that we crash the app on startup to meet the
		// requirements of `google/wire` framework.
		log.Fatal(err)
	}

	s := &HowHearAboutUsItemStorerImpl{
		Logger:     loggerp,
		DbClient:   client,
		Collection: uc,
	}
	return s
}
