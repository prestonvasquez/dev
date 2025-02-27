//go:build WithTransaction

package main

import (
	"context"
	"log"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

func main() {
	client, _ := mongo.Connect()
	client.Disconnect(context.Background())

	sess, err := client.StartSession()
	if err != nil {
		log.Fatalf("failed to start session: %v", err)
	}

	coll := client.Database("db").Collection("coll")
	_, err = sess.WithTransaction(context.Background(), func(ctx context.Context) (interface{}, error) {
		_, err := coll.InsertOne(ctx, bson.D{{"x", 1}})

		return nil, err
	})

	if err != nil {
		log.Fatalf("failed to run transaction: %v", err)
	}
}
