package datastore

import (
	"context"
	"time"

	"log/slog"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (impl OrderStorerImpl) ListByFilter(ctx context.Context, f *OrderPaginationListFilter) (*OrderPaginationListResult, error) {
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
	if !f.InvoiceServiceFeeID.IsZero() {
		filter["invoice_service_fee_id"] = f.InvoiceServiceFeeID
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
	results := []*Order{}
	hasNextPage := false
	for cursor.Next(ctx) {
		document := &Order{}
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

	return &OrderPaginationListResult{
		Results:     results,
		NextCursor:  nextCursor,
		HasNextPage: hasNextPage,
	}, nil
}
