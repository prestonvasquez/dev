package main

import (
	"context"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

func main() {
	client, err := mongo.Connect()
	if err != nil {
		panic(err)
	}

	defer func() { _ = client.Disconnect(context.Background()) }()

	pairs := &mongo.ClientWriteModels{}

	pairs = pairs.AppendInsertOne("db", "", mongo.NewClientInsertOneModel().SetDocument(bson.D{{"x", 1}}))

	_, err = client.BulkWrite(context.Background(), pairs) // Should not panic
	if err != nil {
		panic(err)
	}
}
