package datastore

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// func (impl StaffStorerImpl) ListByInsuranceRequirementID(ctx context.Context, irID primitive.ObjectID) (*StaffPaginationListResult, error) {
// 	f := &StaffPaginationListFilter{
// 		Cursor:                    "",
// 		PageSize:                  1_000_000,
// 		SortField:                 "",
// 		SortOrder:                 0,
// 		InInsuranceRequirementIDs: []primitive.ObjectID{irID},
// 	}
// 	return impl.ListByFilter(ctx, f)
// }

func (impl StaffStorerImpl) ListByHowDidYouHearAboutUsID(ctx context.Context, howDidYouHearAboutUsID primitive.ObjectID) (*StaffPaginationListResult, error) {
	f := &StaffPaginationListFilter{
		Cursor:                 "",
		PageSize:               1_000_000_000, // Max
		SortField:              "",
		SortOrder:              SortOrderAscending,
		HowDidYouHearAboutUsID: howDidYouHearAboutUsID,
	}
	return impl.ListByFilter(ctx, f)
}
