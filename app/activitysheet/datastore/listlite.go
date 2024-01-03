package datastore

import (
	"context"
	"time"

	"log/slog"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ActivitySheetLite struct {
	ID                    primitive.ObjectID `bson:"_id" json:"id"`
	OrderID               primitive.ObjectID `bson:"order_id" json:"order_id"`
	OrderWJID             uint64             `bson:"order_wjid" json:"order_wjid"` // A.K.A. `Workery Job ID`
	AssociateID           primitive.ObjectID `bson:"associate_id" json:"associate_id"`
	AssociateName         string             `bson:"associate_name" json:"associate_name"`
	AssociateLexicalName  string             `bson:"associate_lexical_name" json:"associate_lexical_name"`
	CreatedAt             time.Time          `bson:"created_at" json:"created_at"`
	Status                int8               `bson:"status" json:"status"`
	Type                  int8               `bson:"type_of" json:"type_of"`
	TenantID              primitive.ObjectID `bson:"tenant_id" json:"tenant_id,omitempty"`
	OrderTenantIDWithWJID string             `bson:"order_tenant_id_with_wjid" json:"order_tenant_id_with_wjid"` // TenantIDWithWJID is a combination of `tenancy_id` and `wjid` values written in the following structure `%v_%v`.
}

type ActivitySheetLiteListResult struct {
	Results     []*ActivitySheetLite `json:"results"`
	NextCursor  primitive.ObjectID   `json:"next_cursor"`
	HasNextPage bool                 `json:"has_next_page"`
}

func (impl ActivitySheetStorerImpl) LiteListByFilter(ctx context.Context, f *ActivitySheetPaginationListFilter) (*ActivitySheetPaginationLiteListResult, error) {
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
	if !f.OrderID.IsZero() {
		filter["order_id"] = f.OrderID
	}
	if f.OrderWJID != 0 {
		filter["order_wjid"] = f.OrderWJID
	}
	if !f.AssociateID.IsZero() {
		filter["associate_id"] = f.AssociateID
	}
	if f.ExcludeArchived {
		filter["status"] = bson.M{"$ne": ActivitySheetStatusArchived} // Do not list archived items! This code
	}
	if f.Status != 0 {
		filter["status"] = f.Status
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

	// var results = []*ComicSubmission{}
	// if err = cursor.All(ctx, &results); err != nil {
	// 	panic(err)
	// }

	// Retrieve the documents and check if there is a next page
	results := []*ActivitySheetLite{}
	hasNextPage := false
	for cursor.Next(ctx) {
		document := &ActivitySheetLite{}
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
		nextCursor, err = impl.newPaginatorNextCursor(f, results)
		if err != nil {
			return nil, err
		}
	}

	return &ActivitySheetPaginationLiteListResult{
		Results:     results,
		NextCursor:  nextCursor,
		HasNextPage: hasNextPage,
	}, nil
}
