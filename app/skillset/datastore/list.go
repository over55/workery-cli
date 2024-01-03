package datastore

import (
	"context"
	"time"

	"log/slog"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (impl SkillSetStorerImpl) ListByFilter(ctx context.Context, f *SkillSetPaginationListFilter) (*SkillSetPaginationListResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 12*time.Second)
	defer cancel()

	// Create the paginated filter based on the cursor
	filter, err := impl.newPaginationFilter(f)
	if err != nil {
		return nil, err
	}

	// Add filter conditions to the filter
	if !f.TenantID.IsZero() {
		filter["tenant_id"] = f.TenantID
	}
	if f.Status != 0 {
		filter["status"] = f.Status
	}
	// if !f.CreatedAtGTE.IsZero() {
	// 	filter["created_at"] = bson.M{"$gt": f.CreatedAtGTE} // Add the cursor condition to the filter
	// }
	if f.SearchText != "" {
		filter["sub_category"] = bson.M{"$regex": primitive.Regex{Pattern: f.SearchText, Options: "i"}}
	}

	impl.Logger.Debug("listing filter:",
		slog.Any("filter", filter))

	// Include additional filters for our cursor-based pagination pertaining to sorting and limit.
	options, err := impl.newPaginationOptions(f)
	if err != nil {
		return nil, err
	}

	// // Include Full-text search
	// if f.SearchText != "" {
	// 	filter["$text"] = bson.M{"$search": f.SearchText}
	// 	options.SetProjection(bson.M{"score": bson.M{"$meta": "textScore"}})
	// 	options.SetSort(bson.D{{"score", bson.M{"$meta": "textScore"}}})
	// }

	// Execute the query
	cursor, err := impl.Collection.Find(ctx, filter, options)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	// Retrieve the documents and check if there is a next page
	results := []*SkillSet{}
	hasNextPage := false
	for cursor.Next(ctx) {
		document := &SkillSet{}
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

	return &SkillSetPaginationListResult{
		Results:     results,
		NextCursor:  nextCursor,
		HasNextPage: hasNextPage,
	}, nil
}
