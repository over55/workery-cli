package datastore

import (
	"context"
	"time"

	"log/slog"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	c "github.com/over55/workery-cli/config"
)

const (
	TaskItemStatusActive   = 1
	TaskItemStatusArchived = 2

	TaskItemTypeAssignedAssociate                           = 1
	TaskItemTypeFollowUpDidAssociateAndCustomerAgreedToMeet = 2
	TaskItemTypeFollowUpCustomerSurvey                      = 3 // DEPRECATED
	TaskItemTypeFollowUpDidAssociateAcceptJob               = 4
	TaskItemTypeUpdateOngoingJob                            = 5
	TaskItemTypeFollowUpDidAssociateCompleteJob             = 6
	TaskItemTypeFollowUpDidCustomerReviewAssociateAfterJob  = 7
)

type TaskItem struct {
	ID                                    primitive.ObjectID  `bson:"_id" json:"id"`                                    // 01
	TenantID                              primitive.ObjectID  `bson:"tenant_id" json:"tenant_id,omitempty"`             // 03
	Type                                  int8                `bson:"type" json:"type"`                                 // 04
	Title                                 string              `bson:"title" json:"title"`                               // 05
	Description                           string              `bson:"description" json:"description"`                   // 06
	DueDate                               time.Time           `bson:"due_date" json:"due_date"`                         // 07
	IsClosed                              bool                `bson:"is_closed" json:"is_closed"`                       // 08
	WasPostponed                          bool                `bson:"was_postponed" json:"was_postponed"`               // 09
	ClosingReason                         int8                `bson:"closing_reason" json:"closing_reason"`             // 10
	ClosingReasonOther                    string              `bson:"closing_reason_other" json:"closing_reason_other"` // 11
	OrderID                               primitive.ObjectID  `bson:"order_id" json:"order_id"`                         // 12
	OrderType                             int8                `bson:"order_type" json:"order_type"`                     // 28
	OrderWJID                             uint64              `bson:"order_wjid" json:"order_wjid"`
	OrderTenantIDWithWJID                 string              `bson:"order_tenant_id_with_wjid" json:"order_tenant_id_with_wjid"` // OrderTenantIDWithWJID is a combination of `tenancy_id` and `wjid` values written in the following structure `%v_%v`.
	OrderStartDate                        time.Time           `bson:"order_start_date" json:"order_start_date"`
	OrderDescription                      string              `bson:"order_description" json:"order_description"`
	OrderSkillSets                        []*TaskItemSkillSet `bson:"order_skill_sets" json:"order_skill_sets,omitempty"` // Related
	OrderTags                             []*TaskItemTag      `bson:"order_tags" json:"order_tags,omitempty"`             // Related
	CreatedAt                             time.Time           `bson:"created_at" json:"created_at"`
	CreatedByUserID                       primitive.ObjectID  `bson:"created_by_user_id" json:"created_by_user_id,omitempty"`
	CreatedByUserName                     string              `bson:"created_by_user_name" json:"created_by_user_name"`
	CreatedFromIPAddress                  string              `bson:"created_from_ip_address" json:"created_from_ip_address"`
	ModifiedAt                            time.Time           `bson:"modified_at" json:"modified_at"`
	ModifiedByUserID                      primitive.ObjectID  `bson:"modified_by_user_id" json:"modified_by_user_id,omitempty"`
	ModifiedByUserName                    string              `bson:"modified_by_user_name" json:"modified_by_user_name"`
	ModifiedFromIPAddress                 string              `bson:"modified_from_ip_address" json:"modified_from_ip_address"`
	Status                                int8                `bson:"status" json:"status"` // 20
	OldID                                 uint64              `bson:"old_id" json:"old_id"` // 21
	CustomerID                            primitive.ObjectID  `bson:"customer_id" json:"customer_id"`
	CustomerName                          string              `bson:"customer_name" json:"customer_name,omitempty"`
	CustomerLexicalName                   string              `bson:"customer_lexical_name" json:"customer_lexical_name,omitempty"`
	CustomerGender                        int8                `bson:"customer_gender" json:"customer_gender"`
	CustomerGenderOther                   string              `bson:"customer_gender_other" json:"customer_gender_other"`
	CustomerBirthdate                     time.Time           `bson:"customer_birthdate" json:"customer_birthdate"`
	CustomerEmail                         string              `bson:"customer_email" json:"customer_email,omitempty"`
	CustomerPhone                         string              `bson:"customer_phone" json:"customer_phone,omitempty"`
	CustomerPhoneType                     int8                `bson:"customer_phone_type" json:"customer_phone_type"`
	CustomerPhoneExtension                string              `bson:"customer_phone_extension" json:"customer_phone_extension"`
	CustomerOtherPhone                    string              `bson:"customer_other_phone" json:"customer_other_phone"`
	CustomerOtherPhoneExtension           string              `bson:"customer_other_phone_extension" json:"customer_other_phone_extension"`
	CustomerOtherPhoneType                int8                `bson:"customer_other_phone_type" json:"customer_other_phone_type"`
	CustomerFullAddressWithoutPostalCode  string              `bson:"customer_full_address_without_postal_code" json:"customer_full_address_without_postal_code"`
	CustomerFullAddressURL                string              `bson:"customer_full_address_url" json:"customer_full_address_url"`
	CustomerTags                          []*TaskItemTag      `bson:"customer_tags" json:"customer_tags,omitempty"` // Related
	AssociateID                           primitive.ObjectID  `bson:"associate_id" json:"associate_id"`
	AssociateName                         string              `bson:"associate_name" json:"associate_name,omitempty"`
	AssociateLexicalName                  string              `bson:"associate_lexical_name" json:"associate_lexical_name,omitempty"`
	AssociateGender                       int8                `bson:"associate_gender" json:"associate_gender"`
	AssociateGenderOther                  string              `bson:"associate_gender_other" json:"associate_gender_other"`
	AssociateBirthdate                    time.Time           `bson:"associate_birthdate" json:"associate_birthdate"`
	AssociateEmail                        string              `bson:"associate_email" json:"associate_email,omitempty"`
	AssociatePhone                        string              `bson:"associate_phone" json:"associate_phone,omitempty"`
	AssociatePhoneType                    int8                `bson:"associate_phone_type" json:"associate_phone_type"`
	AssociatePhoneExtension               string              `bson:"associate_phone_extension" json:"associate_phone_extension"`
	AssociateOtherPhone                   string              `bson:"associate_other_phone" json:"associate_other_phone"`
	AssociateOtherPhoneExtension          string              `bson:"associate_other_phone_extension" json:"associate_other_phone_extension"`
	AssociateOtherPhoneType               int8                `bson:"associate_other_phone_type" json:"associate_other_phone_type"`
	AssociateFullAddressWithoutPostalCode string              `bson:"associate_full_address_without_postal_code" json:"associate_full_address_without_postal_code"`
	AssociateFullAddressURL               string              `bson:"associate_full_address_url" json:"associate_full_address_url"`
	AssociateTags                         []*TaskItemTag      `bson:"associate_tags" json:"associate_tags,omitempty"` // Related
}

type TaskItemTag struct {
	ID          primitive.ObjectID `bson:"_id" json:"id"`
	Text        string             `bson:"text" json:"text"`
	Description string             `bson:"description" json:"description"`
	Status      int8               `bson:"status" json:"status"`
}

type TaskItemSkillSet struct {
	ID                    primitive.ObjectID `bson:"_id" json:"id"`
	OrderID               primitive.ObjectID `bson:"order_id" json:"order_id"`
	OrderWJID             uint64             `bson:"order_wjid" json:"order_wjid"`                               // Workery Job ID
	OrderTenantIDWithWJID string             `bson:"order_tenant_id_with_wjid" json:"order_tenant_id_with_wjid"` // OrderTenantIDWithWJID is a combination of `tenancy_id` and `wjid` values written in the following structure `%v_%v`.
	TenantID              primitive.ObjectID `bson:"tenant_id" json:"tenant_id,omitempty"`
	Category              string             `bson:"category" json:"category"`
	SubCategory           string             `bson:"sub_category" json:"sub_category"`
	Description           string             `bson:"description" json:"description"`
	Status                int8               `bson:"status" json:"status"`
	OldID                 uint64             `bson:"old_id" json:"old_id"`
}

// SkillSetIDs is a convinience function which will return an array of skill
// set ID values from the associate.
func (t *TaskItem) SkillSetIDs() []primitive.ObjectID {
	skillSetIDs := make([]primitive.ObjectID, 0)
	for _, ss := range t.OrderSkillSets {
		skillSetIDs = append(skillSetIDs, ss.ID)
	}
	return skillSetIDs
}

type TaskItemListFilter struct {
	// Pagination related.
	Cursor    primitive.ObjectID
	PageSize  int64
	SortField string
	SortOrder int8 // 1=ascending | -1=descending

	// Filter related.
	TenantID        primitive.ObjectID
	CustomerID      primitive.ObjectID
	AssociateID     primitive.ObjectID
	OrderID         primitive.ObjectID
	OrderWJID       uint64 // A.K.A. `Workery Job ID`
	Status          int8
	ExcludeArchived bool
	SearchText      string
	IsClosed        int8
}

type TaskItemListResult struct {
	Results     []*TaskItem        `json:"results"`
	NextCursor  primitive.ObjectID `json:"next_cursor"`
	HasNextPage bool               `json:"has_next_page"`
}

type TaskItemAsSelectOption struct {
	Value primitive.ObjectID `bson:"_id" json:"value"` // Extract from the database `_id` field and output through API as `value`.
	Label string             `bson:"name" json:"label"`
}

// TaskItemStorer Interface for user.
type TaskItemStorer interface {
	Create(ctx context.Context, m *TaskItem) error
	GetByID(ctx context.Context, id primitive.ObjectID) (*TaskItem, error)
	GetByOldID(ctx context.Context, oldID uint64) (*TaskItem, error)
	GetByEmail(ctx context.Context, email string) (*TaskItem, error)
	GetByVerificationCode(ctx context.Context, verificationCode string) (*TaskItem, error)
	CheckIfExistsByEmail(ctx context.Context, email string) (bool, error)
	UpdateByID(ctx context.Context, m *TaskItem) error
	ListByFilter(ctx context.Context, f *TaskItemPaginationListFilter) (*TaskItemPaginationListResult, error)
	ListAsSelectOptionByFilter(ctx context.Context, f *TaskItemListFilter) ([]*TaskItemAsSelectOption, error)
	DeleteByID(ctx context.Context, id primitive.ObjectID) error
	PermanentlyDeleteAllByCustomerID(ctx context.Context, customerID primitive.ObjectID) error
	PermanentlyDeleteAllByAssociateID(ctx context.Context, associateID primitive.ObjectID) error
	CountByFilter(ctx context.Context, f *TaskItemListFilter) (int64, error)
	// //TODO: Add more...
}

type TaskItemStorerImpl struct {
	Logger     *slog.Logger
	DbClient   *mongo.Client
	Collection *mongo.Collection
}

func NewDatastore(appCfg *c.Conf, loggerp *slog.Logger, client *mongo.Client) TaskItemStorer {
	// ctx := context.Background()
	uc := client.Database(appCfg.DB.Name).Collection("task_items")

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

	s := &TaskItemStorerImpl{
		Logger:     loggerp,
		DbClient:   client,
		Collection: uc,
	}
	return s
}
