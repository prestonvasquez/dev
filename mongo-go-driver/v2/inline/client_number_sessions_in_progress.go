//go:build NumberSessionsInProgress

package main

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

func main() {
	client, err := mongo.Connect()
	if err != nil {
		panic(err)
	}

	// Do something with the client
	client.Database("db").Collection("coll").InsertOne(context.Background(), bson.D{{"x", 1}})

	client.Disconnect(context.Background())

	fmt.Println(client.NumberSessionsInProgress())
}
