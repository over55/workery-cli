package datastore

import (
	"context"

	"log/slog"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (impl OrderStorerImpl) Create(ctx context.Context, u *Order) error {
	// DEVELOPER NOTES:
	// According to mongodb documentaiton:
	//     Non-existent Databases and Collections
	//     If the necessary database and collection don't exist when you perform a write operation, the server implicitly creates them.
	//     Source: https://www.mongodb.com/docs/drivers/go/current/usage-examples/insertOne/

	if u.ID == primitive.NilObjectID {
		u.ID = primitive.NewObjectID()
		impl.Logger.Warn("database insert user not included id value, created id now.", slog.Any("id", u.ID))
	}
	// If `public_is` not explicitly set then we implicitly set it.
	if u.WJID == 0 {
		publicID, err := impl.generateWJID(ctx, u.TenantID)
		if err != nil {
			return err
		}
		u.WJID = publicID
	}

	_, err := impl.Collection.InsertOne(ctx, u)

	// check for errors in the insertion
	if err != nil {
		impl.Logger.Error("database insert error", slog.Any("error", err))
	}

	return nil
}

func (impl OrderStorerImpl) generateWJID(ctx context.Context, tenantID primitive.ObjectID) (uint64, error) {
	var publicID uint64
	latest, err := impl.GetLatestOrderByTenantID(ctx, tenantID)
	if err != nil {
		impl.Logger.Error("database get latest order by tenant id error",
			slog.Any("error", err),
			slog.Any("tenant_id", tenantID))
		return 0, err
	}
	if latest == nil {
		impl.Logger.Debug("first order creation detected, setting publicID to value of 1",
			slog.Any("tenant_id", tenantID))
		publicID = 1
	} else {
		publicID = latest.WJID + 1
		impl.Logger.Debug("system generated new order publicID",
			slog.Int("tenant_id", int(publicID)))
	}
	return publicID, nil
}
