package main

import (
	"context"
	"log"

	"go.mongodb.org/mongo-driver/v2/mongo"
)

func main() {
	client, err := mongo.Connect()
	if err != nil {
		panic(err)
	}

	defer func() { _ = client.Disconnect(context.Background()) }()

	pairs := &mongo.ClientWriteModels{}

	_, err = client.BulkWrite(context.Background(), pairs) // Should not panic
	if err != nil {
		log.Fatalf("failed to bulk write: %v", err)
	}
}
