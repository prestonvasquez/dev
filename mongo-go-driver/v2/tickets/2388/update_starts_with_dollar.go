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

	insertOneModel := mongo.NewClientInsertOneModel().SetDocument(bson.D{{"x", 1}})

	pairs = pairs.AppendInsertOne("db", "k", insertOneModel)
	_, err = client.BulkWrite(context.Background(), pairs) // Should not panic
	if err != nil {
		panic(err)
	}

	pairs = &mongo.ClientWriteModels{}

	updateOneModel := mongo.NewClientUpdateOneModel().SetFilter(bson.D{}).SetUpdate(bson.D{{"set", bson.D{{"status", "D"}}}})

	pairs = pairs.AppendUpdateOne("db", "k", updateOneModel)

	_, err = client.BulkWrite(context.Background(), pairs) // Should not panic
	if err != nil {
		panic(err)
	}
}
