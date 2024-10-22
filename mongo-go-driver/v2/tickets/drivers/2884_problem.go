package main

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/event"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func main() {
	poolState := &poolState{}

	opts := options.Client().SetPoolMonitor(poolState.newPoolMonitor())
	client, err := mongo.Connect(opts)
	if err != nil {
		panic(err)
	}

	defer func() { _ = client.Disconnect(context.Background()) }()

	// Create a failpoint that will block an insert for 2s.
	teardown, err := configureBlockingFP(context.Background(), client)
	if err != nil {
		panic(err)
	}

	defer teardown()

	coll := client.Database("db").Collection("coll")

	wg := sync.WaitGroup{}
	wg.Add(10)

	for i := 0; i < 10; i++ {
		go func() {
			defer wg.Done()
			// Create an operation-level context deadline of 1s.
			ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)

			_, _ = coll.InsertOne(ctx, bson.D{{"x", 1}})

			cancel()
		}()
	}

	wg.Wait()

	fmt.Println("size of the bool before disconnecting: ", len(poolState.checkedOut))
}

type poolState struct {
	checkedOut map[int64]struct{}
}

func (ps *poolState) newPoolMonitor() *event.PoolMonitor {
	ps.checkedOut = make(map[int64]struct{})

	return &event.PoolMonitor{
		Event: func(evt *event.PoolEvent) {
			if evt.Type == event.ConnectionCheckedOut {
				ps.checkedOut[evt.ConnectionID] = struct{}{}
			}
			if evt.Type == event.ConnectionClosed {
				delete(ps.checkedOut, evt.ConnectionID)
			}
		},
	}
}

func configureBlockingFP(ctx context.Context, client *mongo.Client) (func(), error) {
	admindb := client.Database("admin")

	// Create a document for the run command that sets a fail command that is always on.
	failCommand := bson.D{
		{"configureFailPoint", "failCommand"},
		{"mode", bson.D{{"times", 10}}},
		{"data",
			bson.D{
				{"blockConnection", true},
				{"blockTimeMS", 750},
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
