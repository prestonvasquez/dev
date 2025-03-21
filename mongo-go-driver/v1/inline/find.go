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
	monitor := event.CommandMonitor{
		Started: func(_ context.Context, cse *event.CommandStartedEvent) {
			if cse.CommandName == "find" {
				fmt.Println("started: ", cse.Command)
			}
		},
	}

	// Set client options
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017").SetMonitor(&monitor).SetTimeout(13 * time.Second)

	// Connect to MongoDB
	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		log.Fatal(err)
	}

	// Check the connection
	err = client.Ping(context.TODO(), nil)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Connected to MongoDB!")

	collection := client.Database("mydatabase").Collection("mycollection")

	// Some sample data
	documents := []interface{}{
		bson.D{{"name", "Alice"}, {"age", 28}},
		bson.D{{"name", "Bob"}, {"age", 34}},
		bson.D{{"name", "Charlie"}, {"age", 25}},
	}

	// Insert data
	insertResult, err := collection.InsertMany(context.TODO(), documents)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Inserted documents: %v\n", insertResult.InsertedIDs)
	//findCtx, findCancel := context.WithTimeout(context.Background(), 15*time.Second)
	//defer findCancel()

	findOpts := options.Find().SetMaxTime(15 * time.Millisecond)

	// Finding multiple documents returns a cursor
	cur, err := collection.Find(context.Background(), bson.D{}, findOpts)
	if err != nil {
		log.Fatal(err)
	}
	defer cur.Close(context.TODO())
	// Iterating through the cursor
	for cur.Next(context.TODO()) {
		var result bson.D
		err := cur.Decode(&result)
		if err != nil {
			log.Fatal(err)
		}
		//fmt.Println(result)
	}

	if err := cur.Err(); err != nil {
		log.Fatal(err)
	}
	// Disconnect from MongoDB
	err = client.Disconnect(context.TODO())
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Disconnected from MongoDB.")
}
