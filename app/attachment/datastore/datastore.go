package datastore

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/exp/slog"

	c "github.com/over55/workery-cli/config"
)

const (
	AttachmentStatusActive    = 1
	AttachmentStatusArchived  = 2
	AttachmentTypeOfCustomer  = 1
	AttachmentTypeOfAssociate = 2
	AttachmentTypeOfOrder     = 3
	AttachmentTypeOfStaff     = 4
)

type Attachment struct {
	ID                    primitive.ObjectID `bson:"_id" json:"id"`
	TenantID              primitive.ObjectID `bson:"tenant_id" json:"tenant_id,omitempty"`
	ObjectKey             string             `bson:"object_key" json:"object_key"`   // 4
	Title                 string             `bson:"title" json:"title"`             // 5
	Description           string             `bson:"description" json:"description"` // 6
	CreatedAt             time.Time          `bson:"created_at" json:"created_at"`
	CreatedByUserID       primitive.ObjectID `bson:"created_by_user_id" json:"created_by_user_id,omitempty"`
	CreatedByUserName     string             `bson:"created_by_user_name" json:"created_by_user_name"`
	CreatedFromIPAddress  string             `bson:"created_from_ip_address" json:"created_from_ip_address"`
	ModifiedAt            time.Time          `bson:"modified_at" json:"modified_at"`
	ModifiedByUserID      primitive.ObjectID `bson:"modified_by_user_id" json:"modified_by_user_id,omitempty"`
	ModifiedByUserName    string             `bson:"modified_by_user_name" json:"modified_by_user_name"`
	ModifiedFromIPAddress string             `bson:"modified_from_ip_address" json:"modified_from_ip_address"`
	AssociateID           primitive.ObjectID `bson:"associate_id" json:"associate_id"` // 14
	AssociateName         string             `bson:"associate_name" json:"associate_name"`
	CustomerID            primitive.ObjectID `bson:"customer_id" json:"customer_id"` //15
	CustomerName          string             `bson:"customer_name" json:"customer_name"`
	StaffID               primitive.ObjectID `bson:"staff_id" json:"staff_id"` // 17
	StaffName             string             `bson:"staff_name" json:"staff_name"`
	OrderID               primitive.ObjectID `bson:"order_id" json:"order_id"` // 18
	Status                int8               `bson:"status" json:"status"`     // 19
	OldID                 uint64             `bson:"old_id" json:"old_id"`     // 20
	TypeOf                int8               `bson:"type_of" json:"type_of"`   // 19
}

type AttachmentListFilter struct {
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

type AttachmentListResult struct {
	Results     []*Attachment      `json:"results"`
	NextCursor  primitive.ObjectID `json:"next_cursor"`
	HasNextPage bool               `json:"has_next_page"`
}

type AttachmentAsSelectOption struct {
	Value primitive.ObjectID `bson:"_id" json:"value"` // Extract from the database `_id` field and output through API as `value`.
	Label string             `bson:"name" json:"label"`
}

// AttachmentStorer Interface for user.
type AttachmentStorer interface {
	Create(ctx context.Context, m *Attachment) error
	GetByID(ctx context.Context, id primitive.ObjectID) (*Attachment, error)
	GetByOldID(ctx context.Context, oldID uint64) (*Attachment, error)
	GetByEmail(ctx context.Context, email string) (*Attachment, error)
	GetByVerificationCode(ctx context.Context, verificationCode string) (*Attachment, error)
	CheckIfExistsByEmail(ctx context.Context, email string) (bool, error)
	UpdateByID(ctx context.Context, m *Attachment) error
	ListByFilter(ctx context.Context, f *AttachmentListFilter) (*AttachmentListResult, error)
	ListAsSelectOptionByFilter(ctx context.Context, f *AttachmentListFilter) ([]*AttachmentAsSelectOption, error)
	DeleteByID(ctx context.Context, id primitive.ObjectID) error
	// //TODO: Add more...
}

type AttachmentStorerImpl struct {
	Logger     *slog.Logger
	DbClient   *mongo.Client
	Collection *mongo.Collection
}

func NewDatastore(appCfg *c.Conf, loggerp *slog.Logger, client *mongo.Client) AttachmentStorer {
	// ctx := context.Background()
	uc := client.Database(appCfg.DB.Name).Collection("attachments")

	// // The following few lines of code will create the index for our app for this
	// // colleciton.
	// indexModel := mongo.IndexModel{
	// 	Keys: bson.D{
	// 		{"text", "text"},
	// 		{"description", "text"},
	// 	},
	// }
	// _, err := uc.Indexes().CreateOne(context.TODO(), indexModel)
	// if err != nil {
	// 	// It is important that we crash the app on startup to meet the
	// 	// requirements of `google/wire` framework.
	// 	log.Fatal(err)
	// }

	s := &AttachmentStorerImpl{
		Logger:     loggerp,
		DbClient:   client,
		Collection: uc,
	}
	return s
}
