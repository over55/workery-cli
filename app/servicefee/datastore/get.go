package datastore

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"log/slog"
)

func (impl ServiceFeeStorerImpl) GetByID(ctx context.Context, id primitive.ObjectID) (*ServiceFee, error) {
	filter := bson.D{{"_id", id}}

	var result ServiceFee
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

func (impl ServiceFeeStorerImpl) GetByOldID(ctx context.Context, oldID uint64) (*ServiceFee, error) {
	filter := bson.D{{"old_id", oldID}}

	var result ServiceFee
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

func (impl ServiceFeeStorerImpl) GetByEmail(ctx context.Context, email string) (*ServiceFee, error) {
	filter := bson.D{{"email", email}}

	var result ServiceFee
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

func (impl ServiceFeeStorerImpl) GetByVerificationCode(ctx context.Context, verificationCode string) (*ServiceFee, error) {
	filter := bson.D{{"email_verification_code", verificationCode}}

	var result ServiceFee
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
