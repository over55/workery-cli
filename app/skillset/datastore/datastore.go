package datastore

import (
	"context"
	"log"
	"time"

	"log/slog"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	c "github.com/over55/workery-cli/config"
)

const (
	SkillSetStatusActive   = 1
	SkillSetStatusArchived = 2
	SkillSetRoleRoot       = 1
	SkillSetRoleRetailer   = 2
	SkillSetRoleCustomer   = 3
)

type SkillSet struct {
	ID                    primitive.ObjectID              `bson:"_id" json:"id"`
	TenantID              primitive.ObjectID              `bson:"tenant_id" json:"tenant_id,omitempty"`
	Category              string                          `bson:"category" json:"category"`
	SubCategory           string                          `bson:"sub_category" json:"sub_category"`
	Description           string                          `bson:"description" json:"description"`
	Status                int8                            `bson:"status" json:"status"`
	PublicID              uint64                          `bson:"public_id" json:"public_id"`
	InsuranceRequirements []*SkillSetInsuranceRequirement `bson:"insurance_requirements" json:"insurance_requirements,omitempty"` // Reference
	CreatedAt             time.Time                       `bson:"created_at" json:"created_at"`
	CreatedByUserID       primitive.ObjectID              `bson:"created_by_user_id" json:"created_by_user_id,omitempty"`
	CreatedByUserName     string                          `bson:"created_by_user_name" json:"created_by_user_name"`
	CreatedFromIPAddress  string                          `bson:"created_from_ip_address" json:"created_from_ip_address"`
	ModifiedAt            time.Time                       `bson:"modified_at" json:"modified_at"`
	ModifiedByUserID      primitive.ObjectID              `bson:"modified_by_user_id" json:"modified_by_user_id,omitempty"`
	ModifiedByUserName    string                          `bson:"modified_by_user_name" json:"modified_by_user_name"`
	ModifiedFromIPAddress string                          `bson:"modified_from_ip_address" json:"modified_from_ip_address"`
}

// SkillSetInsuranceRequirement structure is a copy of `InsuranceRequirement` with extra `SkillSetID` field.
type SkillSetInsuranceRequirement struct {
	SkillSetID  primitive.ObjectID `bson:"skill_set_id" json:"skill_set_id"`
	TenantID    primitive.ObjectID `bson:"tenant_id" json:"tenant_id"`
	ID          primitive.ObjectID `bson:"_id" json:"id"`
	PublicID    uint64             `bson:"public_id" json:"public_id"`
	Name        string             `bson:"name" json:"name"`
	Description string             `bson:"description" json:"description"`
	Status      int8               `bson:"status" json:"status"`
}

type SkillSetListFilter struct {
	// Pagination related.
	Cursor    primitive.ObjectID
	PageSize  int64
	SortField string
	SortOrder int8 // 1=ascending | -1=descending

	// Filter related.
	TenantID        primitive.ObjectID
	Role            int8
	Status          int8
	UUIDs           []string
	ExcludeArchived bool
	SearchText      string
	FirstName       string
	LastName        string
	Email           string
	Phone           string
	CreatedAtGTE    time.Time
}

type SkillSetListResult struct {
	Results     []*SkillSet        `json:"results"`
	NextCursor  primitive.ObjectID `json:"next_cursor"`
	HasNextPage bool               `json:"has_next_page"`
}

type SkillSetAsSelectOption struct {
	Value primitive.ObjectID `bson:"_id" json:"value"` // Extract from the database `_id` field and output through API as `value`.
	Label string             `bson:"sub_category" json:"label"`
}

// SkillSetStorer Interface for user.
type SkillSetStorer interface {
	Create(ctx context.Context, m *SkillSet) error
	GetByID(ctx context.Context, id primitive.ObjectID) (*SkillSet, error)
	GetByPublicID(ctx context.Context, oldID uint64) (*SkillSet, error)
	GetLatestByTenantID(ctx context.Context, tenantID primitive.ObjectID) (*SkillSet, error)
	CheckIfExistsByEmail(ctx context.Context, email string) (bool, error)
	UpdateByID(ctx context.Context, m *SkillSet) error
	ListByFilter(ctx context.Context, f *SkillSetPaginationListFilter) (*SkillSetPaginationListResult, error)
	ListAsSelectOptionByFilter(ctx context.Context, f *SkillSetListFilter) ([]*SkillSetAsSelectOption, error)
	ListByTenantID(ctx context.Context, tenantID primitive.ObjectID) (*SkillSetPaginationListResult, error)
	DeleteByID(ctx context.Context, id primitive.ObjectID) error
}

type SkillSetStorerImpl struct {
	Logger     *slog.Logger
	DbClient   *mongo.Client
	Collection *mongo.Collection
}

func NewDatastore(appCfg *c.Conf, loggerp *slog.Logger, client *mongo.Client) SkillSetStorer {
	// ctx := context.Background()
	uc := client.Database(appCfg.DB.Name).Collection("skill_sets")

	_, err := uc.Indexes().CreateMany(context.TODO(), []mongo.IndexModel{
		{Keys: bson.D{{Key: "tenant_id", Value: 1}}},
		{Keys: bson.D{{Key: "public_id", Value: -1}}},
		{Keys: bson.D{{Key: "status", Value: 1}}},
		{Keys: bson.D{
			{"category", "text"},
			{"sub_category", "text"},
			{"description", "text"},
		}},
	})
	if err != nil {
		// It is important that we crash the app on startup to meet the
		// requirements of `google/wire` framework.
		log.Fatal(err)
	}

	s := &SkillSetStorerImpl{
		Logger:     loggerp,
		DbClient:   client,
		Collection: uc,
	}
	return s
}
