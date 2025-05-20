package main

import (
	"context"
	"log"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.opentelemetry.io/contrib/instrumentation/go.mongodb.org/mongo-driver/v2/mongo/otelmongo"
)

func main() {
	telhooks := otelmongo.NewTelemetryHooks()

	clientOpts := options.Client()
	clientOpts.TelemetryHooks = telhooks

	client, err := mongo.Connect(clientOpts)
	if err != nil {
		log.Fatalf("failed to connect to MongoDB: %v", err)
	}

	defer func() {
		if err := client.Disconnect(context.Background()); err != nil {
			log.Fatalf("failed to disconnect from MongoDB: %v", err)
		}
	}()

	coll := client.Database("test").Collection("example")

	_, err = coll.InsertOne(context.Background(), map[string]string{"name": "example"})
	if err != nil {
		log.Fatalf("failed to insert document: %v", err)
	}
}
