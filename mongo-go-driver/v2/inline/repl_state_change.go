package main

import (
	"context"
	"fmt"
	"log"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/event"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func main() {
	opts := options.Client().SetMonitor(monitor())

	client, err := mongo.Connect(opts)
	if err != nil {
		panic(err)
	}

	defer func() { _ = client.Disconnect(context.Background()) }()

	teardown, err := configureBlockingFP(context.Background(), client)
	if err != nil {
		panic(err)
	}

	defer teardown()

	coll := client.Database("db").Collection("coll")

	_, err = coll.InsertOne(context.Background(), bson.D{})
	if err != nil {
		panic(err)
	}
}

func configureBlockingFP(ctx context.Context, client *mongo.Client) (func(), error) {
	admindb := client.Database("admin")

	// Create a document for the run command that sets a fail command that is always on.
	failCommand := bson.D{
		{"configureFailPoint", "failCommand"},
		{"mode", bson.D{{"times", 1}}},
		{"data",
			bson.D{
				{"failCommands", bson.A{"insert"}},
				{"errorCode", 11602},
				{"closeConnection", false},
			},
		},
	}

	err := admindb.RunCommand(ctx, failCommand).Err()
	if err != nil {
		return func() {}, err
	}

	return func() {
		doc := bson.D{
			{"configureFailPoint", "failCommand"},
			{"mode", "off"},
		}
		err = admindb.RunCommand(ctx, doc).Err()
		if err != nil {
			log.Fatalf("could not disable fail point command: %v", err)
		}
	}, nil
}

func monitor() *event.CommandMonitor {
	return &event.CommandMonitor{
		Started: func(_ context.Context, cse *event.CommandStartedEvent) {
			if cse.CommandName == "insert" {
				fmt.Println("started: ", cse)
			}
		},
		Succeeded: func(_ context.Context, cse *event.CommandSucceededEvent) {
			if cse.CommandName == "insert" {
				fmt.Println("succeeded: ", cse)
			}
		},
		Failed: func(_ context.Context, cfe *event.CommandFailedEvent) {
			if cfe.CommandName == "insert" {
				fmt.Println("failed: ", cfe)
			}
		},
	}
}
