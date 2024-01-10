package datastore

import (
	"context"
	"time"

	"log/slog"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (impl AssociateStorerImpl) ListByFilter(ctx context.Context, f *AssociatePaginationListFilter) (*AssociatePaginationListResult, error) {
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
	if !f.HowDidYouHearAboutUsID.IsZero() {
		filter["how_did_you_hear_about_us_id"] = f.TenantID
	}
	if f.Role > 0 {
		filter["role"] = f.Role
	}
	if f.FirstName != "" {
		filter["first_name"] = bson.M{"$regex": primitive.Regex{Pattern: f.FirstName, Options: "i"}}
	}
	if f.LastName != "" {
		filter["last_name"] = bson.M{"$regex": primitive.Regex{Pattern: f.LastName, Options: "i"}}
	}
	if f.Email != "" {
		filter["email"] = bson.M{"$regex": primitive.Regex{Pattern: f.Email, Options: "i"}}
	}
	if f.Phone != "" {
		filter["phone"] = f.Phone
	}
	if f.Status != 0 {
		filter["status"] = f.Status
	}
	if !f.CreatedAtGTE.IsZero() {
		filter["created_at"] = bson.M{"$gt": f.CreatedAtGTE} // Add the cursor condition to the filter
	}
	if f.Type != 0 {
		filter["type"] = f.Type
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
	if len(f.InInsuranceRequirementIDs) > 0 {
		filter["insurance_requirements._id"] = bson.M{"$in": f.InInsuranceRequirementIDs}
	}
	if len(f.AllInsuranceRequirementIDs) > 0 {
		filter["insurance_requirements._id"] = bson.M{"$all": f.AllInsuranceRequirementIDs}
	}
	if len(f.InVehicleTypeIDs) > 0 {
		filter["vehicle_types._id"] = bson.M{"$in": f.InVehicleTypeIDs}
	}
	if len(f.AllVehicleTypeIDs) > 0 {
		filter["vehicle_types._id"] = bson.M{"$all": f.AllVehicleTypeIDs}
	}

	// Create a slice to store conditions
	var conditions []bson.M

	// Add filter conditions to the slice
	if !f.CommercialInsuranceExpiryDateGTE.IsZero() {
		conditions = append(conditions, bson.M{"commercial_insurance_expiry_date": bson.M{"$gte": f.CommercialInsuranceExpiryDateGTE}})
	}
	if !f.CommercialInsuranceExpiryDateGT.IsZero() {
		conditions = append(conditions, bson.M{"commercial_insurance_expiry_date": bson.M{"$gt": f.CommercialInsuranceExpiryDateGT}})
	}
	if !f.CommercialInsuranceExpiryDateLTE.IsZero() {
		conditions = append(conditions, bson.M{"commercial_insurance_expiry_date": bson.M{"$lte": f.CommercialInsuranceExpiryDateLTE}})
	}
	if !f.CommercialInsuranceExpiryDateLT.IsZero() {
		conditions = append(conditions, bson.M{"commercial_insurance_expiry_date": bson.M{"$lt": f.CommercialInsuranceExpiryDateLT}})
	}

	if !f.PoliceCheckGTE.IsZero() {
		conditions = append(conditions, bson.M{"police_check": bson.M{"$gte": f.PoliceCheckGTE}})
	}
	if !f.PoliceCheckGT.IsZero() {
		conditions = append(conditions, bson.M{"police_check": bson.M{"$gt": f.PoliceCheckGT}})
	}
	if !f.PoliceCheckLTE.IsZero() {
		conditions = append(conditions, bson.M{"police_check": bson.M{"$lte": f.PoliceCheckLTE}})
	}
	if !f.PoliceCheckLT.IsZero() {
		conditions = append(conditions, bson.M{"police_check": bson.M{"$lt": f.PoliceCheckLT}})
	}

	// Combine conditions with $and operator
	if len(conditions) > 0 {
		filter["$and"] = conditions
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
	results := []*Associate{}
	hasNextPage := false
	for cursor.Next(ctx) {
		document := &Associate{}
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
		nextCursor, err = impl.newPaginatorNextCursorForFull(f, results)
		if err != nil {
			return nil, err
		}
	}

	return &AssociatePaginationListResult{
		Results:     results,
		NextCursor:  nextCursor,
		HasNextPage: hasNextPage,
	}, nil
}
