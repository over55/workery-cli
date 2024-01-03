package datastore

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (impl TaskItemStorerImpl) ListByCustomerID(ctx context.Context, customerID primitive.ObjectID) (*TaskItemPaginationListResult, error) {
	f := &TaskItemPaginationListFilter{
		Cursor:     "",
		PageSize:   1_000_00,
		SortField:  "", // Setting this empty to ignore any sorting.
		SortOrder:  SortOrderAscending,
		CustomerID: customerID,
	}
	res, err := impl.ListByFilter(ctx, f)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (impl TaskItemStorerImpl) ListByAssociateID(ctx context.Context, associateID primitive.ObjectID) (*TaskItemPaginationListResult, error) {
	f := &TaskItemPaginationListFilter{
		Cursor:      "",
		PageSize:    1_000_00,
		SortField:   "", // Setting this empty to ignore any sorting.
		SortOrder:   SortOrderAscending,
		AssociateID: associateID,
	}
	res, err := impl.ListByFilter(ctx, f)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (impl TaskItemStorerImpl) ListByOrderID(ctx context.Context, orderID primitive.ObjectID) (*TaskItemPaginationListResult, error) {
	f := &TaskItemPaginationListFilter{
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

func (impl TaskItemStorerImpl) ListByOrderWJID(ctx context.Context, orderWJID uint64) (*TaskItemPaginationListResult, error) {
	f := &TaskItemPaginationListFilter{
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
