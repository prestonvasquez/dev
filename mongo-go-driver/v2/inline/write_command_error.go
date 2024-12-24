package main

import (
	"context"
	"fmt"
	"log"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/x/mongo/driver"
)

func main() {
	client, err := mongo.Connect()
	if err != nil {
		log.Fatalf("failed to connect to server: %v", err)
	}

	defer func() {
		if err := client.Disconnect(context.Background()); err != nil {
			log.Fatalf("failed to disconnect from server: %v", err)
		}
	}()

	fpTeardown, err := configureBlockingFP(context.Background(), client)
	if err != nil {
		log.Fatalf("failed to config failpoint: %v", err)
	}

	defer fpTeardown()

	coll := client.Database("db").Collection("coll")

	defer func() {
		if err := coll.Drop(context.Background()); err != nil {
			log.Fatalf("failed to drop connection: %v", err)
		}
	}()

	_, err = coll.InsertOne(context.Background(), bson.D{{"x", 1}})
	if err == nil {
		log.Fatal("expected write to fail")
	}

	driverErr, ok := err.(driver.Error)
	if !ok {
		log.Fatal("error not driver error")
	}

	fmt.Println(driverErr)
}

func configureBlockingFP(ctx context.Context, client *mongo.Client) (func(), error) {
	admindb := client.Database("admin")

	// Create a document for the run command that sets a fail command that is always on.
	failCommand := bson.D{
		{"configureFailPoint", "failCommand"},
		{"mode", bson.D{{"times", 2}}},
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
