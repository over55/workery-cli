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
	CommentStatusActive   = 1
	CommentStatusArchived = 100
	BelongsToCustomer     = 1
	BelongsToAssociate    = 2
	BelongsToOrder        = 3
	BelongsToStaff        = 4
)

type Comment struct {
	ID                    primitive.ObjectID `bson:"_id" json:"id"`
	BelongsTo             int8               `bson:"belongs_to" json:"belongs_to"`
	CustomerID            primitive.ObjectID `bson:"customer_id" json:"customer_id,omitempty"`
	CustomerName          string             `bson:"customer_name" json:"customer_name"`
	AssociateID           primitive.ObjectID `bson:"associate_id" json:"associate_id,omitempty"`
	AssociateName         string             `bson:"associate_name" json:"associate_name"`
	OrderID               primitive.ObjectID `bson:"order_id" json:"order_id,omitempty"`
	OrderWJID             uint64             `bson:"order_wjid" json:"order_wjid"`
	OrderTenantIDWithWJID string             `bson:"order_tenant_id_with_wjid" json:"order_tenant_id_with_wjid"` // OrderTenantIDWithWJID is a combination of `tenancy_id` and `wjid` values written in the following structure `%v_%v`.
	StaffID               primitive.ObjectID `bson:"staff_id" json:"staff_id,omitempty"`
	StaffName             string             `bson:"staff_name" json:"staff_name"`
	TenantID              primitive.ObjectID `bson:"tenant_id" json:"tenant_id,omitempty"`
	CreatedAt             time.Time          `bson:"created_at" json:"created_at"`
	CreatedByUserID       primitive.ObjectID `bson:"created_by_user_id" json:"created_by_user_id,omitempty"`
	CreatedByUserName     string             `bson:"created_by_user_name" json:"created_by_user_name"`
	CreatedFromIPAddress  string             `bson:"created_from_ip_address" json:"created_from_ip_address"`
	ModifiedAt            time.Time          `bson:"modified_at" json:"modified_at"`
	ModifiedByUserID      primitive.ObjectID `bson:"modified_by_user_id" json:"modified_by_user_id,omitempty"`
	ModifiedByUserName    string             `bson:"modified_by_user_name" json:"modified_by_user_name"`
	ModifiedFromIPAddress string             `bson:"modified_from_ip_address" json:"modified_from_ip_address"`
	Content               string             `bson:"content" json:"content"`
	Status                int8               `bson:"status" json:"status"`
	OldID                 uint64             `bson:"old_id" json:"old_id"`
}

type CommentListFilter struct {
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

type CommentListResult struct {
	Results     []*Comment         `json:"results"`
	NextCursor  primitive.ObjectID `json:"next_cursor"`
	HasNextPage bool               `json:"has_next_page"`
}

type CommentAsSelectOption struct {
	Value primitive.ObjectID `bson:"_id" json:"value"` // Extract from the database `_id` field and output through API as `value`.
	Label string             `bson:"name" json:"label"`
}

// CommentStorer Interface for user.
type CommentStorer interface {
	Create(ctx context.Context, m *Comment) error
	GetByID(ctx context.Context, id primitive.ObjectID) (*Comment, error)
	GetByOldID(ctx context.Context, oldID uint64) (*Comment, error)
	GetByEmail(ctx context.Context, email string) (*Comment, error)
	GetByVerificationCode(ctx context.Context, verificationCode string) (*Comment, error)
	CheckIfExistsByEmail(ctx context.Context, email string) (bool, error)
	UpdateByID(ctx context.Context, m *Comment) error
	ListByFilter(ctx context.Context, f *CommentListFilter) (*CommentListResult, error)
	ListAsSelectOptionByFilter(ctx context.Context, f *CommentListFilter) ([]*CommentAsSelectOption, error)
	DeleteByID(ctx context.Context, id primitive.ObjectID) error
	// //TODO: Add more...
}

type CommentStorerImpl struct {
	Logger     *slog.Logger
	DbClient   *mongo.Client
	Collection *mongo.Collection
}

func NewDatastore(appCfg *c.Conf, loggerp *slog.Logger, client *mongo.Client) CommentStorer {
	// ctx := context.Background()
	uc := client.Database(appCfg.DB.Name).Collection("comments")

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

	s := &CommentStorerImpl{
		Logger:     loggerp,
		DbClient:   client,
		Collection: uc,
	}
	return s
}
