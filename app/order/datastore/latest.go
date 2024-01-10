package datastore

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (impl OrderStorerImpl) GetLatestOrderByTenantID(ctx context.Context, tenantID primitive.ObjectID) (*Order, error) {
	filter := bson.D{{"tenant_id", tenantID}}
	opts := options.Find().SetSort(bson.D{{"wjid", -1}}).SetLimit(1)

	var order Order
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

	return nil, nil
}

func (impl OrderStorerImpl) GetLatestCommentByOrderID(ctx context.Context, orderID primitive.ObjectID) (*OrderComment, error) {
	filter := bson.M{"order_id": orderID}
	options := options.FindOne().SetSort(bson.M{"created_at": -1})

	var comment OrderComment
	err := impl.Collection.FindOne(ctx, filter, options).Decode(&comment)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil // No documents found, return nil without an error
		}
		return nil, err
	}

	return &comment, nil
}

func GetLatestComment(order *Order) *OrderComment {
	if order == nil || len(order.Comments) == 0 {
		return nil
	}

	latestComment := order.Comments[0]

	for _, comment := range order.Comments[1:] {
		if comment.CreatedAt.After(latestComment.CreatedAt) {
			latestComment = comment
		}
	}

	return latestComment
}
