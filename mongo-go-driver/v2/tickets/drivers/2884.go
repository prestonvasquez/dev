package main

import (
	"context"
	"fmt"
	"log"
	"sort"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/event"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func main() {
	opts := options.Client().SetPoolMonitor(newPoolMonitor()).SetMonitor(newMonitor()).SetMaxPoolSize(1)

	client, err := mongo.Connect(opts)
	if err != nil {
		panic(err)
	}

	ctx := context.Background()

	defer func() {
		_ = client.Disconnect(ctx)
	}()

	durations := make([]time.Duration, 100)

	for i := range durations {
		durations[i], err = runTest(ctx, client)
		if err != nil {
			panic(err)
		}
	}

	fmt.Println("median op time w/ blocking 750 ms and 50 ms maxTimeMS: %v", calcMedian(durations))
}

func calcMedian(durations []time.Duration) time.Duration {
	durationsInt := make([]int64, len(durations))
	for i, d := range durations {
		durationsInt[i] = int64(d)
	}

	sort.Slice(durationsInt, func(i, j int) bool {
		return durationsInt[i] < durationsInt[j]
	})

	n := len(durationsInt)
	if n%2 == 1 {
		return time.Duration(durationsInt[n/2])
	}

	middle1 := durationsInt[n/2-1]
	middle2 := durationsInt[n/2]

	return time.Duration(middle1+middle2) / 2
}

func runTest(ctx context.Context, client *mongo.Client) (time.Duration, error) {
	// Create a failpoint that will block an insert for 2s.
	teardown, err := configureBlockingFP(ctx, client)
	if err != nil {
		return 0, err
	}

	defer teardown()

	// Create a collection to operate over.
	coll := client.Database("db").Collection("coll")

	// Create an operation-level context deadline of 1s.
	ctx, cancel := context.WithTimeout(ctx, 50*time.Millisecond)
	defer cancel()

	st := time.Now()
	// Execute an insert operation.
	_, _ = coll.InsertOne(ctx, bson.D{{"x", 1}})

	return time.Since(st), nil
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
