package main

import (
	"context"
	"log"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

func main() {
	// Connect to MongoDB
	client, err := mongo.Connect()
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err = client.Disconnect(context.TODO()); err != nil {
			log.Fatal(err)
		}
	}()

	// Create a document with a field large enough to reach 16 MB
	largeString := make([]byte, 16777216+100) // Adjust size to account for BSON overhead
	for i := range largeString {
		largeString[i] = 'a'
	}

	document := bson.D{{"largeField", string(largeString)}}

	pairs := &mongo.ClientWriteModels{}

	pairs = pairs.AppendInsertOne("db", "x", mongo.NewClientInsertOneModel().SetDocument(document))

	_, err = client.BulkWrite(context.Background(), pairs) // Should not panic
	if err != nil {
		panic(err)
	}
}
