package datastore

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (impl ActivitySheetStorerImpl) ListByOrderID(ctx context.Context, orderID primitive.ObjectID) (*ActivitySheetPaginationListResult, error) {
	f := &ActivitySheetPaginationListFilter{
		Cursor:    "",
		PageSize:  1_000_00,
		SortField: "", // Setting this empty to ignore any sorting.
		SortOrder: SortOrderAscending,
		OrderID:   orderID,
	}
	res, err := impl.ListByFilter(ctx, f)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (impl ActivitySheetStorerImpl) ListByOrderWJID(ctx context.Context, orderWJID uint64) (*ActivitySheetPaginationListResult, error) {
	f := &ActivitySheetPaginationListFilter{
		Cursor:    "",
		PageSize:  1_000_00,
		SortField: "", // Setting this empty to ignore any sorting.
		SortOrder: SortOrderAscending,
		OrderWJID: orderWJID,
	}
	res, err := impl.ListByFilter(ctx, f)
	if err != nil {
		return nil, err
	}
	return res, nil
}
