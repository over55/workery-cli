package datastore

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (impl AssociateStorerImpl) ListByInsuranceRequirementID(ctx context.Context, irID primitive.ObjectID) (*AssociatePaginationListResult, error) {
	f := &AssociatePaginationListFilter{
		Cursor:                    "",
		PageSize:                  1_000_000,
		SortField:                 "",
		SortOrder:                 0,
		InInsuranceRequirementIDs: []primitive.ObjectID{irID},
	}
	return impl.ListByFilter(ctx, f)
}

func (impl AssociateStorerImpl) ListByHowDidYouHearAboutUsID(ctx context.Context, howDidYouHearAboutUsID primitive.ObjectID) (*AssociatePaginationListResult, error) {
	f := &AssociatePaginationListFilter{
		Cursor:                 "",
		PageSize:               1_000_000_000, // Max
		SortField:              "",
		SortOrder:              0,
		HowDidYouHearAboutUsID: howDidYouHearAboutUsID,
	}
	return impl.ListByFilter(ctx, f)
}
