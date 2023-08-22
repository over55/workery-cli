package datastore

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/exp/slog"
)

func (impl PrivateImageStorerImpl) GetByID(ctx context.Context, id primitive.ObjectID) (*PrivateImage, error) {
	filter := bson.D{{"_id", id}}

	var result PrivateImage
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

func (impl PrivateImageStorerImpl) GetByOldID(ctx context.Context, oldID uint64) (*PrivateImage, error) {
	filter := bson.D{{"old_id", oldID}}

	var result PrivateImage
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

func (impl PrivateImageStorerImpl) GetByEmail(ctx context.Context, email string) (*PrivateImage, error) {
	filter := bson.D{{"email", email}}

	var result PrivateImage
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

func (impl PrivateImageStorerImpl) GetByVerificationCode(ctx context.Context, verificationCode string) (*PrivateImage, error) {
	filter := bson.D{{"email_verification_code", verificationCode}}

	var result PrivateImage
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
