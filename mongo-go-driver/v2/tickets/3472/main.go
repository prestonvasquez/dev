package main

import (
	"context"
	"fmt"
	"log"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

func main() {
	client, err := mongo.Connect()
	if err != nil {
		log.Fatalf("failed to connect to MongoDB: %v", err)
	}

	defer func() {
		if err := client.Disconnect(context.Background()); err != nil {
			log.Fatalf("failed to disconnect from MongoDB: %v", err)
		}
	}()

	coll := client.Database("test").Collection("myCollection")

	_ = coll.Drop(context.Background())

	type myStruct struct {
		V []int8
	}

	_, err = coll.InsertOne(context.Background(), bson.D{{"v", bson.NewVector([]int8{1, 2, 3})}})
	if err != nil {
		log.Fatalf("failed to insert document: %v", err)
	}

	result := coll.FindOne(context.Background(), bson.D{})
	if result.Err() != nil {
		log.Fatalf("failed to find document: %v", result.Err())
	}

	ms := myStruct{}
	if err := result.Decode(&ms); err != nil {
		log.Fatalf("failed to decode document: %v", err)
	}

	fmt.Println(ms.V)
}
