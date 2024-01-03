package datastore

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (impl OrderStorerImpl) ListByCustomerID(ctx context.Context, customerID primitive.ObjectID) (*OrderPaginationListResult, error) {
	f := &OrderPaginationListFilter{
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

func (impl OrderStorerImpl) ListByAssociateID(ctx context.Context, associateID primitive.ObjectID) (*OrderPaginationListResult, error) {
	f := &OrderPaginationListFilter{
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

func (impl OrderStorerImpl) ListByServiceFeeID(ctx context.Context, serviceFeeID primitive.ObjectID) (*OrderPaginationListResult, error) {
	f := &OrderPaginationListFilter{
		Cursor:              "",
		PageSize:            1_000_00,
		SortField:           "", // Setting this empty to ignore any sorting.
		SortOrder:           SortOrderAscending,
		InvoiceServiceFeeID: serviceFeeID,
	}
	res, err := impl.ListByFilter(ctx, f)
	if err != nil {
		return nil, err
	}
	return res, nil
}
