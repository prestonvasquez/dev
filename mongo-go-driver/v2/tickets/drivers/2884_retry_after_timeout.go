package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/event"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type eventTracker struct {
	commandStartedEvents int
}

func (et *eventTracker) Started(_ context.Context, e *event.CommandStartedEvent) {
	if e.CommandName == "insert" {
		et.commandStartedEvents++
	}
}

func (et *eventTracker) ClearEvents() {
	et.commandStartedEvents = 0
}

func main() {
	eventTracker := &eventTracker{}

	opts := options.Client().SetMonitor(&event.CommandMonitor{
		Started: eventTracker.Started,
	})

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

	if _, err := coll.InsertOne(context.Background(), bson.D{}); err != nil {
		panic(err)
	}

	eventTracker.ClearEvents()

	for i := 0; i < 2; i++ {
		// Run 50 operations, each with a timeout of 50ms. Expect
		// them to all return a timeout error because the failpoint
		// blocks find operations for 500ms. Run 50 to increase the
		// probability that an operation will time out in a way that
		// can cause a retry.
		ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)

		_, err := coll.InsertOne(ctx, bson.D{})
		cancel()

		if !errors.Is(err, context.DeadlineExceeded) {
			panic("err is not deadline exceeded")
		}

		if !mongo.IsTimeout(err) {
			panic("err is not a timeout")
		}

		fmt.Println("commandStartedEvents count: ", eventTracker.commandStartedEvents)
		eventTracker.ClearEvents()

		// Assert that each operation reported exactly one command
		// started events, which means the operation did not retry
		// after the context timeout.
		//evts := mt.GetAllStartedEvents()
		//require.Len(mt,
		//	mt.GetAllStartedEvents(),
		//	1,
		//	"expected exactly 1 command started event per operation, but got %d after %d iterations",
		//	len(evts),
		//	i)
		//mt.ClearEvents()
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
