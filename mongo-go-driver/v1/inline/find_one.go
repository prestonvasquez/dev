package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/event"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	// Set up a context
	ctx := context.TODO()

	monitor := event.CommandMonitor{
		Started: func(_ context.Context, cse *event.CommandStartedEvent) {
			if cse.CommandName == "find" {
				fmt.Println(cse.Command)
			}
		},
	}

	// Connect to MongoDB
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017").SetMonitor(&monitor)
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatal(err)
	}

	// Ensure the connection is established
	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Connected to MongoDB")

	// Access a collection
	collection := client.Database("mydatabase").Collection("mycollection")

	// Define the filter for the `FindOne`
	filter := bson.M{"fieldName": "someValue"}

	// Define a variable where the result will be decoded
	var result bson.M

	findOneCtx, findOneCancel := context.WithTimeout(ctx, 10*time.Millisecond)
	defer findOneCancel()

	opts := options.FindOne().SetMaxTime(5 * time.Millisecond)
	err = collection.FindOne(findOneCtx, filter, opts).Decode(&result)
	if err == mongo.ErrNoDocuments {
		fmt.Println("No document was found")
	} else if err != nil {
		log.Fatal(err)
	} else {
		fmt.Println("Document found:", result)
	}

	// Clean up and disconnect
	if err = client.Disconnect(ctx); err != nil {
		log.Fatal(err)
	}
	fmt.Println("Disconnected from MongoDB")
}
