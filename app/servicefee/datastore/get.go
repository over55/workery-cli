package datastore

import (
	"context"

	"log/slog"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
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

func (impl ServiceFeeStorerImpl) GetByPublicID(ctx context.Context, oldID uint64) (*ServiceFee, error) {
	filter := bson.D{{"public_id", oldID}}

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

func (impl ServiceFeeStorerImpl) GetLatestByTenantID(ctx context.Context, tenantID primitive.ObjectID) (*ServiceFee, error) {
	filter := bson.D{{"tenant_id", tenantID}}
	opts := options.Find().SetSort(bson.D{{"public_id", -1}}).SetLimit(1)

	var order ServiceFee
	cursor, err := impl.Collection.Find(context.Background(), filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	if cursor.Next(context.Background()) {
		err := cursor.Decode(&order)
		if err != nil {
			return nil, err
		}
		return &order, nil
	}

	return nil, mongo.ErrNoDocuments
}
