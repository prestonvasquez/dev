package main

import (
	"context"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/mongo/writeconcern"
)

func main() {
	client, err := mongo.Connect()
	if err != nil {
		panic(err)
	}

	defer func() { _ = client.Disconnect(context.Background()) }()

	pairs := &mongo.ClientWriteModels{}

	insertOneModel := mongo.NewClientInsertOneModel().SetDocument(bson.D{{"x", 1}})

	opts := options.ClientBulkWrite().SetWriteConcern(writeconcern.Unacknowledged()).SetOrdered(false)

	pairs = pairs.AppendInsertOne("db", "k", insertOneModel)
	_, err = client.BulkWrite(context.Background(), pairs, opts) // Should not panic
	if err != nil {
		panic(err)
	}
}
