package datastore

import (
	"context"
	"time"

	"log/slog"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type OrderLite struct {
	ID                   primitive.ObjectID `bson:"_id" json:"id"`
	WJID                 uint64             `bson:"wjid" json:"wjid"` // A.K.A. `Workery Job ID`
	CustomerID           primitive.ObjectID `bson:"customer_id" json:"customer_id"`
	CustomerName         string             `bson:"customer_name" json:"customer_name,omitempty"`
	CustomerLexicalName  string             `bson:"customer_lexical_name" json:"customer_lexical_name,omitempty"`
	CustomerGender       int8               `bson:"customer_gender" json:"customer_gender"`
	CustomerGenderOther  string             `bson:"customer_gender_other" json:"customer_gender_other"`
	CustomerBirthdate    time.Time          `bson:"customer_birthdate" json:"customer_birthdate"`
	AssociateID          primitive.ObjectID `bson:"associate_id" json:"associate_id"`
	AssociateName        string             `bson:"associate_name" json:"associate_name,omitempty"`
	AssociateLexicalName string             `bson:"associate_lexical_name" json:"associate_lexical_name,omitempty"`
	AssociateGender      int8               `bson:"associate_gender" json:"associate_gender"`
	AssociateGenderOther string             `bson:"associate_gender_other" json:"associate_gender_other"`
	AssociateBirthdate   time.Time          `bson:"associate_birthdate" json:"associate_birthdate"`
	TenantID             primitive.ObjectID `bson:"tenant_id" json:"tenant_id,omitempty"`
	Description          string             `bson:"description" json:"description"`
	AssignmentDate       time.Time          `bson:"assignment_date" json:"assignment_date"`
	IsOngoing            bool               `bson:"is_ongoing" json:"is_ongoing"`
	IsHomeSupportService bool               `bson:"is_home_support_service" json:"is_home_support_service"`
	StartDate            time.Time          `bson:"start_date" json:"start_date"`
	CompletionDate       time.Time          `bson:"completion_date" json:"completion_date"`
	Status               int8               `bson:"status" json:"status"`
	Type                 int8               `bson:"type" json:"type"`
	CreatedAt            time.Time          `bson:"created_at" json:"created_at"`
	CreatedByUserID      primitive.ObjectID `bson:"created_by_user_id" json:"created_by_user_id,omitempty"`
	CreatedByUserName    string             `bson:"created_by_user_name" json:"created_by_user_name"`
	ModifiedAt           time.Time          `bson:"modified_at" json:"modified_at"`
	ModifiedByUserID     primitive.ObjectID `bson:"modified_by_user_id" json:"modified_by_user_id,omitempty"`
	ModifiedByUserName   string             `bson:"modified_by_user_name" json:"modified_by_user_name"`
}

func (impl OrderStorerImpl) LiteListByFilter(ctx context.Context, f *OrderPaginationListFilter) (*OrderPaginationLiteListResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 12*time.Second)
	defer cancel()

	filter, err := impl.newPaginationFilter(f)
	if err != nil {
		return nil, err
	}

	// Add filter conditions to the filter
	if !f.TenantID.IsZero() {
		filter["tenant_id"] = f.TenantID
	}
	if !f.CustomerID.IsZero() {
		filter["customer_id"] = f.CustomerID
	}
	if !f.AssociateID.IsZero() {
		filter["associate_id"] = f.AssociateID
	}
	if f.ExcludeArchived {
		filter["status"] = bson.M{"$ne": OrderStatusArchived} // Do not list archived items! This code
	}
	if f.Status != 0 {
		filter["status"] = f.Status
	}
	if f.Type != 0 {
		filter["type"] = f.Type
	}
	if !f.ModifiedByUserID.IsZero() {
		filter["modified_by_user_id"] = f.ModifiedByUserID
	}
	if f.CustomerFirstName != "" {
		filter["customer_first_name"] = bson.M{"$regex": primitive.Regex{Pattern: f.CustomerFirstName, Options: "i"}}
	}
	if f.CustomerLastName != "" {
		filter["customer_last_name"] = bson.M{"$regex": primitive.Regex{Pattern: f.CustomerLastName, Options: "i"}}
	}
	if f.CustomerEmail != "" {
		filter["customer_email"] = bson.M{"$regex": primitive.Regex{Pattern: f.CustomerEmail, Options: "i"}}
	}
	if f.CustomerPhone != "" {
		filter["customer_phone"] = f.CustomerPhone
	}
	if f.AssociateFirstName != "" {
		filter["associate_first_name"] = bson.M{"$regex": primitive.Regex{Pattern: f.AssociateFirstName, Options: "i"}}
	}
	if f.AssociateLastName != "" {
		filter["associate_last_name"] = bson.M{"$regex": primitive.Regex{Pattern: f.AssociateLastName, Options: "i"}}
	}
	if f.AssociateEmail != "" {
		filter["associate_email"] = bson.M{"$regex": primitive.Regex{Pattern: f.AssociateEmail, Options: "i"}}
	}
	if f.AssociatePhone != "" {
		filter["associate_phone"] = f.AssociatePhone
	}
	if len(f.InSkillSetIDs) > 0 {
		filter["skill_sets._id"] = bson.M{"$in": f.InSkillSetIDs}
	}
	if len(f.AllSkillSetIDs) > 0 {
		filter["skill_sets._id"] = bson.M{"$all": f.AllSkillSetIDs}
	}
	if len(f.InTagIDs) > 0 {
		filter["tags._id"] = bson.M{"$in": f.InTagIDs}
	}
	if len(f.AllTagIDs) > 0 {
		filter["tags._id"] = bson.M{"$all": f.AllTagIDs}
	}

	impl.Logger.Debug("listing filter:",
		slog.Any("filter", filter))

	// Include additional filters for our cursor-based pagination pertaining to sorting and limit.
	options, err := impl.newPaginationOptions(f)
	if err != nil {
		return nil, err
	}

	// Include Full-text search
	if f.SearchText != "" {
		filter["$text"] = bson.M{"$search": f.SearchText}
		options.SetProjection(bson.M{"score": bson.M{"$meta": "textScore"}})
		options.SetSort(bson.D{{"score", bson.M{"$meta": "textScore"}}})
	}

	// Execute the query
	cursor, err := impl.Collection.Find(ctx, filter, options)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	// Retrieve the documents and check if there is a next page
	results := []*OrderLite{}
	hasNextPage := false
	for cursor.Next(ctx) {
		document := &OrderLite{}
		if err := cursor.Decode(document); err != nil {
			return nil, err
		}
		results = append(results, document)
		// Stop fetching documents if we have reached the desired page size
		if int64(len(results)) >= f.PageSize {
			hasNextPage = true
			break
		}
	}

	// Get the next cursor and encode it
	var nextCursor string
	if hasNextPage {
		nextCursor, err = impl.newPaginatorLiteNextCursor(f, results)
		if err != nil {
			return nil, err
		}
	}

	return &OrderPaginationLiteListResult{
		Results:     results,
		NextCursor:  nextCursor,
		HasNextPage: hasNextPage,
	}, nil
}
