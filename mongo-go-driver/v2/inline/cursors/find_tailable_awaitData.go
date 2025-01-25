package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func main() {
	client, err := mongo.Connect(options.Client().SetHeartbeatInterval(500 * time.Millisecond))
	if err != nil {
		log.Fatal("failed to connect to server: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	defer func() {
		if err = client.Disconnect(ctx); err != nil {
			log.Fatal("failed to disconnect client: %v", err)
		}
	}()

	time.Sleep(2 * time.Second)

	// Specify the database and collection
	db := client.Database("exampleDB")
	collection := db.Collection("exampleCappedCollection")

	// Ensure the collection is capped (tailable cursors only work on capped collections)
	//err = setupCappedCollection(ctx, collection)
	//if err != nil {
	//	log.Fatal(err)
	//}

	// Configure the tailable cursor
	findOptions := options.Find()
	findOptions.SetCursorType(options.TailableAwait)
	findOptions.SetBatchSize(1)
	findOptions.SetMaxAwaitTime(500 * time.Millisecond)

	// Start watching for changes in the capped collection
	cursor, err := collection.Find(context.Background(), bson.M{}, findOptions)
	if err != nil {
		log.Fatal("Error creating tailable cursor:", err)
	}
	defer cursor.Close(ctx)

	ctx, cancel = context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	ctx = context.WithValue(ctx, "latency_context", true)

	st := time.Now()
	var docs []bson.Raw
	if err := cursor.All(ctx, &docs); err != nil {
		log.Fatalf("cursor failed: %v", err)
	}

	fmt.Println("elapsed: ", time.Since(st))
}
