//go:build rs2

package main

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// What happens if we run a Find command with a very large timout, but have a
// client with a very small timeout?
//
// Result: operation-level timeout is honored
func main() {
	client, err := mongo.Connect(options.Client().SetTimeout(1 * time.Second))
	if err != nil {
		panic(err)
	}

	defer func() { _ = client.Disconnect(context.Background()) }()

	// Set the first fail point which will be the "previous" error.
	fpclose, err := configureBlockingFP(context.Background(), client, "find")
	if err != nil {
		panic(err)
	}

	defer fpclose()

	// Execute a find operation with a very long timeout
	ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
	defer cancel()

	coll := client.Database("db").Collection("coll")

	startTime := time.Now()
	_, err = coll.Find(ctx, bson.D{})
	if err != nil {
		panic(err)
	}

	log.Println("time elapsed: ", time.Since(startTime))
}

func configureBlockingFP(ctx context.Context, client *mongo.Client, cmd string) (func(), error) {
	admindb := client.Database("admin")

	// Create a document for the run command that sets a fail command that is always on.
	failCommand := bson.D{
		{"configureFailPoint", "failCommand"},
		{"mode", bson.D{{"times", 1}}},
		{"data",
			bson.D{
				{"blockConnection", true},
				{"blockTimeMS", 10_000}, // 10 seconds
				{"failCommands", bson.A{cmd}},
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
