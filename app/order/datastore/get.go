package datastore

import (
	"context"

	"log/slog"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func (impl OrderStorerImpl) GetByID(ctx context.Context, id primitive.ObjectID) (*Order, error) {
	filter := bson.D{{"_id", id}}

	var result Order
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

func (impl OrderStorerImpl) GetByWJID(ctx context.Context, wjID uint64) (*Order, error) {
	filter := bson.D{{"wjid", wjID}}

	var result Order
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

func (impl OrderStorerImpl) GetByEmail(ctx context.Context, email string) (*Order, error) {
	filter := bson.D{{"email", email}}

	var result Order
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

func (impl OrderStorerImpl) GetByVerificationCode(ctx context.Context, verificationCode string) (*Order, error) {
	filter := bson.D{{"email_verification_code", verificationCode}}

	var result Order
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
