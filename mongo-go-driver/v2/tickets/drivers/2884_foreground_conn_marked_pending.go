package main

import (
	"context"
	"errors"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/event"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// Ensure that the foreground gets marked as "pending" and checked back in
// if a CSOT error occurs.

func main() {
	opts := options.Client().SetMaxPoolSize(1).SetPoolMonitor(newPoolMonitor()).SetMonitor(newMonitor())

	client, err := mongo.Connect(opts)
	if err != nil {
		panic(err)
	}

	defer func() {
		log.Println("---- Client Disconnecting ----")
		_ = client.Disconnect(context.Background())
	}()

	teardown, err := configureBlockingFP(context.Background(), client)
	if err != nil {
		panic(err)
	}

	defer teardown()

	coll := client.Database("db").Collection("coll")

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	_, err = coll.InsertOne(ctx, bson.D{})
	if !errors.Is(err, context.DeadlineExceeded) {
		panic(err)
	}
}

func configureBlockingFP(ctx context.Context, client *mongo.Client) (func(), error) {
	admindb := client.Database("admin")

	// Create a document for the run command that sets a fail command that is always on.
	failCommand := bson.D{
		{"configureFailPoint", "failCommand"},
		{"mode", "alwaysOn"},
		{"data",
			bson.D{
				{"blockConnection", true},
				{"blockTimeMS", 500},
				{"failCommands", bson.A{"insert"}},
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

func newMonitor() *event.CommandMonitor {
	return &event.CommandMonitor{
		Started: func(ctx context.Context, cse *event.CommandStartedEvent) {
			if cse.CommandName == "insert" {
				log.Printf("command started: %+v\n", cse)
			}
		},
		Failed: func(ctx context.Context, cfe *event.CommandFailedEvent) {
			if cfe.CommandName == "insert" {
				log.Printf("command failed: %+v\n", cfe)
			}
		},
	}
}

func newPoolMonitor() *event.PoolMonitor {
	return &event.PoolMonitor{
		Event: func(evt *event.PoolEvent) {
			if evt.Type == event.ConnectionClosed {
				log.Printf("connection closed: %+v\n", evt)
			}
		},
	}
}
