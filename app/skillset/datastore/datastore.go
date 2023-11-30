package datastore

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"log/slog"

	c "github.com/over55/workery-cli/config"
)

const (
	SkillSetStatusActive   = 1
	SkillSetStatusArchived = 100
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
	OldID                 uint64                          `bson:"old_id" json:"old_id"`
	InsuranceRequirements []*SkillSetInsuranceRequirement `bson:"insurance_requirements" json:"insurance_requirements,omitempty"` // Reference
}

// SkillSetInsuranceRequirement structure is a copy of `InsuranceRequirement` with extra `SkillSetID` field.
type SkillSetInsuranceRequirement struct {
	SkillSetID  primitive.ObjectID `bson:"skill_set_id" json:"skill_set_id"`
	TenantID    primitive.ObjectID `bson:"tenant_id" json:"tenant_id"`
	ID          primitive.ObjectID `bson:"_id" json:"id"`
	OldID       uint64             `bson:"old_id" json:"old_id"`
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
	OrganizationID  primitive.ObjectID
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
	Label string             `bson:"name" json:"label"`
}

// SkillSetStorer Interface for user.
type SkillSetStorer interface {
	Create(ctx context.Context, m *SkillSet) error
	GetByID(ctx context.Context, id primitive.ObjectID) (*SkillSet, error)
	GetByOldID(ctx context.Context, oldID uint64) (*SkillSet, error)
	CheckIfExistsByEmail(ctx context.Context, email string) (bool, error)
	UpdateByID(ctx context.Context, m *SkillSet) error
	ListByFilter(ctx context.Context, f *SkillSetListFilter) (*SkillSetListResult, error)
	ListAsSelectOptionByFilter(ctx context.Context, f *SkillSetListFilter) ([]*SkillSetAsSelectOption, error)
	ListAllRootStaff(ctx context.Context) (*SkillSetListResult, error)
	ListAllRetailerStaffForOrganizationID(ctx context.Context, organizationID primitive.ObjectID) (*SkillSetListResult, error)
	DeleteByID(ctx context.Context, id primitive.ObjectID) error
	// //TODO: Add more...
}

type SkillSetStorerImpl struct {
	Logger     *slog.Logger
	DbClient   *mongo.Client
	Collection *mongo.Collection
}

func NewDatastore(appCfg *c.Conf, loggerp *slog.Logger, client *mongo.Client) SkillSetStorer {
	// ctx := context.Background()
	uc := client.Database(appCfg.DB.Name).Collection("skill_sets")

	// The following few lines of code will create the index for our app for this
	// colleciton.
	indexModel := mongo.IndexModel{
		Keys: bson.D{
			{"category", "text"},
			{"sub_category", "text"},
			{"description", "text"},
		},
	}
	_, err := uc.Indexes().CreateOne(context.TODO(), indexModel)
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
