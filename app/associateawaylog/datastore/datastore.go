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
	AssociateAwayLogStatusActive   = 1
	AssociateAwayLogStatusArchived = 2
)

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

type AssociateAwayLogListFilter struct {
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

type AssociateAwayLogListResult struct {
	Results     []*AssociateAwayLog `json:"results"`
	NextCursor  primitive.ObjectID  `json:"next_cursor"`
	HasNextPage bool                `json:"has_next_page"`
}

type AssociateAwayLogAsSelectOption struct {
	Value primitive.ObjectID `bson:"_id" json:"value"` // Extract from the database `_id` field and output through API as `value`.
	Label string             `bson:"name" json:"label"`
}

// AssociateAwayLogStorer Interface for user.
type AssociateAwayLogStorer interface {
	Create(ctx context.Context, m *AssociateAwayLog) error
	GetByID(ctx context.Context, id primitive.ObjectID) (*AssociateAwayLog, error)
	GetByOldID(ctx context.Context, oldID uint64) (*AssociateAwayLog, error)
	GetByEmail(ctx context.Context, email string) (*AssociateAwayLog, error)
	GetByVerificationCode(ctx context.Context, verificationCode string) (*AssociateAwayLog, error)
	CheckIfExistsByEmail(ctx context.Context, email string) (bool, error)
	UpdateByID(ctx context.Context, m *AssociateAwayLog) error
	ListByFilter(ctx context.Context, f *AssociateAwayLogListFilter) (*AssociateAwayLogListResult, error)
	ListAsSelectOptionByFilter(ctx context.Context, f *AssociateAwayLogListFilter) ([]*AssociateAwayLogAsSelectOption, error)
	DeleteByID(ctx context.Context, id primitive.ObjectID) error
	// //TODO: Add more...
}

type AssociateAwayLogStorerImpl struct {
	Logger     *slog.Logger
	DbClient   *mongo.Client
	Collection *mongo.Collection
}

func NewDatastore(appCfg *c.Conf, loggerp *slog.Logger, client *mongo.Client) AssociateAwayLogStorer {
	// ctx := context.Background()
	uc := client.Database(appCfg.DB.Name).Collection("associate_away_log")

	// The following few lines of code will create the index for our app for this
	// colleciton.
	indexModel := mongo.IndexModel{
		Keys: bson.D{
			{"associate_name", "text"},
		},
	}
	_, err := uc.Indexes().CreateOne(context.TODO(), indexModel)
	if err != nil {
		// It is important that we crash the app on startup to meet the
		// requirements of `google/wire` framework.
		log.Fatal(err)
	}

	s := &AssociateAwayLogStorerImpl{
		Logger:     loggerp,
		DbClient:   client,
		Collection: uc,
	}
	return s
}
