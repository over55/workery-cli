package datastore

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/exp/slog"
)

func (impl HowHearAboutUsItemStorerImpl) GetByID(ctx context.Context, id primitive.ObjectID) (*HowHearAboutUsItem, error) {
	filter := bson.D{{"_id", id}}

	var result HowHearAboutUsItem
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

func (impl HowHearAboutUsItemStorerImpl) GetByOldID(ctx context.Context, oldID uint64) (*HowHearAboutUsItem, error) {
	filter := bson.D{{"old_id", oldID}}

	var result HowHearAboutUsItem
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

func (impl HowHearAboutUsItemStorerImpl) GetByEmail(ctx context.Context, email string) (*HowHearAboutUsItem, error) {
	filter := bson.D{{"email", email}}

	var result HowHearAboutUsItem
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

func (impl HowHearAboutUsItemStorerImpl) GetByVerificationCode(ctx context.Context, verificationCode string) (*HowHearAboutUsItem, error) {
	filter := bson.D{{"email_verification_code", verificationCode}}

	var result HowHearAboutUsItem
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
