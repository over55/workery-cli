package mongodb

import (
	"context"
	"log"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"log/slog"

	c "github.com/over55/workery-cli/config"
)

func NewStorage(appCfg *c.Conf) *mongo.Client {
	log.Println("storage mongodb initializing...", slog.String("URI", appCfg.DB.URI))
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(appCfg.DB.URI))
	if err != nil {
		log.Fatal(err)
	}

	// The MongoDB client provides a Ping() method to tell you if a MongoDB database has been found and connected.
	if err := client.Ping(context.TODO(), readpref.Primary()); err != nil {
		log.Fatal(err)
	}
	log.Println("storage mongodb initialized successfully")
	return client
}
