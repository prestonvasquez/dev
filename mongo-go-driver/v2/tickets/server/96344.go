package main

import (
	"context"
	"errors"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func createBlockFP(client *mongo.Client, cmd string, blockTime, iter int) (func(), error) {
	admindb := client.Database("admin")

	// Create a document for the run command that sets a fail command that is always on.
	failCommand := bson.D{
		{Key: "configureFailPoint", Value: "failCommand"},
		{Key: "mode", Value: bson.D{{"times", iter}}},
		{Key: "data", Value: bson.D{
			{Key: "blockConnection", Value: true},
			{Key: "blockTimeMS", Value: blockTime},
			{Key: "failCommands", Value: bson.A{cmd}},
		}},
	}

	err := admindb.RunCommand(context.Background(), failCommand).Err()
	if err != nil {
		return func() {}, err
	}

	return func() {
		doc := bson.D{
			{Key: "configureFailPoint", Value: "failCommand"},
			{Key: "mode", Value: "off"},
		}

		err = admindb.RunCommand(context.Background(), doc).Err()
		if err != nil {
			log.Fatalf("could not disable fail point command: %v", err)
		}
	}, nil
}

func main() {
	// Set up MongoDB client options
	opts := options.Client().SetMaxPoolSize(1)

	client, err := mongo.Connect(opts)
	if err != nil {
		log.Fatalf("failed to connect to server: %v", err)
	}
	defer func() {
		_ = client.Disconnect(context.Background())
	}()

	// Create a failpoint that will block 1 time for 1000ms.
	teardown, err := createBlockFP(client, "insert", 1000, 1)
	if err != nil {
		log.Fatalf("failed to create blocking failpoint: %v", err)
	}
	defer teardown()

	db := client.Database("db")

	cmd := bson.D{
		{"insert", "coll"},
		{"documents", bson.A{bson.D{{"x", 1}}}},
		{"maxTimeMS", 50},
	}

	st := time.Now()
	res := db.RunCommand(context.Background(), cmd)
	if err := res.Err(); !errors.Is(err, context.DeadlineExceeded) {
		log.Fatalf("wrong error: %v", err)
	}

	elapsed := time.Since(st)

	if elapsed < 1*time.Second {
		log.Fatalf("expected failpoint to block for 1 second, but blocked for %v", elapsed)
	}

	log.Println("passed")
}
