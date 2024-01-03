package datastore

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (impl CustomerStorerImpl) ListByHowDidYouHearAboutUsID(ctx context.Context, howDidYouHearAboutUsID primitive.ObjectID) (*CustomerPaginationListResult, error) {
	f := &CustomerPaginationListFilter{
		Cursor:                 "",
		PageSize:               1_000_000_000, // Max
		SortField:              "",
		SortOrder:              0,
		HowDidYouHearAboutUsID: howDidYouHearAboutUsID,
	}
	return impl.ListByFilter(ctx, f)
}
