package main

import (
	"context"
	"fmt"
	"log"
	"sync"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/event"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// Can two concurrent operations check out the same connection?
func main() {
	opts := options.Client().SetPoolMonitor(newPoolMonitor()).SetMaxPoolSize(1).SetMonitor(newCommandMonitor())

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

	coll := client.Database("test").Collection("coll")

	fmt.Println("starting")
	wg := sync.WaitGroup{}
	wg.Add(2)

	go func() {
		defer wg.Done()
		coll.InsertOne(context.Background(), bson.D{{"x", 1}})
	}()

	go func() {
		defer wg.Done()
		coll.InsertOne(context.Background(), bson.D{{"x", 1}})
	}()

	wg.Wait()
	fmt.Println("finished")
}

func configureBlockingFP(ctx context.Context, client *mongo.Client) (func(), error) {
	admindb := client.Database("admin")

	// Create a document for the run command that sets a fail command that is always on.
	failCommand := bson.D{
		{"configureFailPoint", "failCommand"},
		{"mode", bson.D{{"times", 2}}},
		{"data",
			bson.D{
				{"blockConnection", true},
				{"blockTimeMS", 10_000}, // 10 seconds
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

func newPoolMonitor() *event.PoolMonitor {
	return &event.PoolMonitor{
		Event: func(pe *event.PoolEvent) {
			switch pe.Type {
			case event.ConnectionCheckedOut:
				fmt.Println("checked out: ", pe)
			}
		},
	}
}

func newCommandMonitor() *event.CommandMonitor {
	return &event.CommandMonitor{
		Started: func(_ context.Context, cse *event.CommandStartedEvent) {
			if cse.CommandName == "insert" {
				fmt.Println("started: ", cse)
			}
		},
	}
}
