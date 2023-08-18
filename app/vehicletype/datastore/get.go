package datastore

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/exp/slog"
)

func (impl VehicleTypeStorerImpl) GetByID(ctx context.Context, id primitive.ObjectID) (*VehicleType, error) {
	filter := bson.D{{"_id", id}}

	var result VehicleType
	err := impl.Collection.FindOne(ctx, filter).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			// This error means your query did not match any documents.
			return nil, nil
		}
		impl.Logger.Error("database get by user id error", slog.Any("error", err))
		return nil, err
	}
	return &result, nil
}

func (impl VehicleTypeStorerImpl) GetByOldID(ctx context.Context, oldID uint64) (*VehicleType, error) {
	filter := bson.D{{"old_id", oldID}}

	var result VehicleType
	err := impl.Collection.FindOne(ctx, filter).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			// This error means your query did not match any documents.
			return nil, nil
		}
		impl.Logger.Error("database get by user id error", slog.Any("error", err))
		return nil, err
	}
	return &result, nil
}

func (impl VehicleTypeStorerImpl) GetByEmail(ctx context.Context, email string) (*VehicleType, error) {
	filter := bson.D{{"email", email}}

	var result VehicleType
	err := impl.Collection.FindOne(ctx, filter).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			// This error means your query did not match any documents.
			return nil, nil
		}
		impl.Logger.Error("database get by email error", slog.Any("error", err))
		return nil, err
	}
	return &result, nil
}

func (impl VehicleTypeStorerImpl) GetByVerificationCode(ctx context.Context, verificationCode string) (*VehicleType, error) {
	filter := bson.D{{"email_verification_code", verificationCode}}

	var result VehicleType
	err := impl.Collection.FindOne(ctx, filter).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			// This error means your query did not match any documents.
			return nil, nil
		}
		impl.Logger.Error("database get by verification code error", slog.Any("error", err))
		return nil, err
	}
	return &result, nil
}
