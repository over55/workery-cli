package datastore

import (
	"context"
	"log/slog"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (impl ActivitySheetStorerImpl) CountByFilter(ctx context.Context, f *ActivitySheetPaginationListFilter) (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 12*time.Second)
	defer cancel()

	// Create the filter based on the cursor
	filter := bson.M{}

	// Add filter conditions to the filter
	if !f.TenantID.IsZero() {
		filter["tenant_id"] = f.TenantID
	}
	if !f.OrderID.IsZero() {
		filter["order_id"] = f.OrderID
	}
	if !f.AssociateID.IsZero() {
		filter["associate_id"] = f.AssociateID
	}
	if f.OrderWJID != 0 {
		filter["order_wjid"] = f.OrderWJID
	}
	if f.ExcludeArchived {
		filter["status"] = bson.M{"$ne": ActivitySheetStatusArchived} // Do not list archived items! This code
	}
	if f.Status != 0 {
		filter["status"] = f.Status
	}

	impl.Logger.Debug("counting w/ filter:",
		slog.Any("filter", filter))

	// Use the CountDocuments method to count the matching documents.
	count, err := impl.Collection.CountDocuments(ctx, filter)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (impl ActivitySheetStorerImpl) CountByLast30DaysForAssociateID(ctx context.Context, associateID primitive.ObjectID) (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 12*time.Second)
	defer cancel()

	// Calculate the date for 30 days ago
	thirtyDaysAgo := time.Now().AddDate(0, 0, -30)

	// Create the filter based on the cursor
	filter := bson.M{
		"associate_id": associateID,
		"date":         bson.M{"$gte": thirtyDaysAgo},
	}

	impl.Logger.Debug("counting w/ filter:",
		slog.Any("filter", filter))

	// Use the CountDocuments method to count the matching documents.
	count, err := impl.Collection.CountDocuments(ctx, filter)
	if err != nil {
		return 0, err
	}

	return count, nil
}
