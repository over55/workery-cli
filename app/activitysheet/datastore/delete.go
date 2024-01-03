package datastore

import (
	"context"
	"log/slog"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (impl ActivitySheetStorerImpl) DeleteByID(ctx context.Context, id primitive.ObjectID) error {
	_, err := impl.Collection.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return err
	}
	return nil
}

func (impl *ActivitySheetStorerImpl) PermanentlyDeleteAllByAssociateID(ctx context.Context, associateID primitive.ObjectID) error {
	f := &ActivitySheetPaginationListFilter{
		PageSize:    1_000_000,
		SortField:   "_id",
		SortOrder:   1, // 1=ascending | -1=descending
		AssociateID: associateID,
	}
	res, err := impl.ListByFilter(ctx, f)
	if err != nil {
		impl.Logger.Error("database list by associate id error", slog.Any("error", err))
		return err
	}
	for _, a := range res.Results {
		if err := impl.DeleteByID(ctx, a.ID); err != nil {
			impl.Logger.Error("database delete by id error",
				slog.Any("error", err))
			return err
		}
	}

	return nil
}
