package datastore

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (impl AttachmentStorerImpl) ListByOrderID(ctx context.Context, orderID primitive.ObjectID) (*AttachmentListResult, error) {
	f := &AttachmentListFilter{
		Cursor:    primitive.NilObjectID,
		PageSize:  1_000_000,
		SortField: "id",
		SortOrder: OrderAscending,
		OrderID:   orderID,
	}
	return impl.ListByFilter(ctx, f)
}

func (impl AttachmentStorerImpl) ListByOrderWJID(ctx context.Context, orderWJID uint64) (*AttachmentListResult, error) {
	f := &AttachmentListFilter{
		Cursor:    primitive.NilObjectID,
		PageSize:  1_000_000,
		SortField: "id",
		SortOrder: OrderAscending,
		OrderWJID: orderWJID,
	}
	return impl.ListByFilter(ctx, f)
}

func (impl AttachmentStorerImpl) ListByType(ctx context.Context, typeOf int8) (*AttachmentListResult, error) {
	f := &AttachmentListFilter{
		Cursor:    primitive.NilObjectID,
		PageSize:  1_000_000,
		SortField: "id",
		SortOrder: OrderAscending,
		Type:      typeOf,
	}
	return impl.ListByFilter(ctx, f)
}
