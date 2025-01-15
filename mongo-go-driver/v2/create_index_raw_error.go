package main

import (
	"context"
	"errors"
	"fmt"
	"log"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func main() {
	clientOpts := options.Client().ApplyURI("mongodb://localhost:27017")
	client, err := mongo.Connect(clientOpts)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(context.TODO())

	// Set up the failpoint to simulate index creation error
	failPointCmd := bson.D{
		{"configureFailPoint", "failCommand"},
		{"mode", bson.D{{"times", 1}}},
		{"data",
			bson.D{
				{"failCommands", bson.A{"createIndexes"}},
				{"errorCode", 11000},
			},
		},
	}

	// Run the command to set up the failpoint
	err = client.Database("admin").RunCommand(context.TODO(), failPointCmd).Err()
	if err != nil {
		log.Fatalf("failed to run failpoint command: %v", err)
	}
	fmt.Println("Failpoint set for index creation.")

	// Try to create an index
	coll := client.Database("testdb").Collection("testcoll")
	indexModel := mongo.IndexModel{
		Keys:    map[string]interface{}{"field": 1},
		Options: options.Index().SetName("field_index"),
	}

	_, err = coll.Indexes().CreateOne(context.TODO(), indexModel)

	var wexceptionErr *mongo.WriteException
	if errors.As(err, &wexceptionErr) {
		fmt.Printf("Index creation failed with error: %+v\n", wexceptionErr)
	} else {
		fmt.Println("Index created successfully.")
	}

	// Disable the failpoint after testing
	failPointDisableCmd := map[string]interface{}{
		"configureFailPoint": "failCommand",
		"mode":               "off",
	}

	err = client.Database("admin").RunCommand(context.TODO(), failPointDisableCmd).Err()
	if err != nil {
		log.Fatalf("failed to disable failpoint: %v", err)
	}
	fmt.Println("Failpoint disabled.")
}
