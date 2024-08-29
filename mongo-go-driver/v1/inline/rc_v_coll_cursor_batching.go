//go:build i2

package main

import (
	"context"
	"log"
	"os"
	"strconv"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/event"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Test the difference between batching cursors with runnCommand and batching
// them using colleciton methods. Specifically under an exceeded timeout.

func main() {
	if len(os.Args) != 2 {
		log.Fatalf("expected a timoutMS arg")
	}

	timeoutMS, err := strconv.Atoi(os.Args[1])
	if err != nil {
		log.Fatalf("failed to parse timeoutMS: %v", err)
	}

	timeout := time.Duration(timeoutMS) * time.Millisecond

	const (
		numBatches = 10
		numRecords = 100
	)

	opts := options.Client().SetMonitor(newCommandMonitor())

	client, err := mongo.Connect(context.Background(), opts)
	if err != nil {
		panic(err)
	}

	defer func() { _ = client.Disconnect(context.Background()) }()

	// Insert a bunch of test data to test the cursors with.
	db := client.Database("testdb")

	coll := db.Collection("test")
	defer func() { _ = coll.Drop(context.Background()) }()

	var docs []any
	for i := 1; i <= numRecords; i++ {
		docs = append(docs, bson.D{{"name", i}})
	}

	if _, err := coll.InsertMany(context.Background(), docs); err != nil {
		panic(err)
	}

	// First try the RunCursorCommand.
	cmd := bson.D{{"find", "test"}, {"batchSize", numBatches}}

	rcCtx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	rcCur, err := db.RunCommandCursor(rcCtx, cmd)
	if err != nil {
		log.Fatalf("error doing run command: %v\n", err)
	}

	defer rcCur.Close(rcCtx)

	//rcCur.SetBatchSize(numBatches)

	rcBatchCount, err := countBatches(rcCtx, rcCur, numBatches)
	if err != nil {
		panic(err)
	}

	log.Println("rc batch count: ", rcBatchCount)

	// Now try with a Find command.
	findCtx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	findCur, err := coll.Find(findCtx, bson.D{}, options.Find().SetBatchSize(numBatches))
	if err != nil {
		panic(err)
	}

	findBatchCount, err := countBatches(findCtx, findCur, numBatches)
	if err != nil {
		log.Fatalf("error doing find command: %v\n", err)
	}

	log.Println("find batch count: ", findBatchCount)

}

func newCommandMonitor() *event.CommandMonitor {
	return &event.CommandMonitor{
		Started: func(_ context.Context, evt *event.CommandStartedEvent) {
			if evt.CommandName == "getMore" {
				log.Println(evt.Command)
			}
		},
		Succeeded: func(_ context.Context, evt *event.CommandSucceededEvent) {
			if evt.CommandName == "getMore" {
				log.Println("succeeded: ", evt.Reply)
			}
		},
	}
}

// countBatches iterates through the cursor, counting the number of batches
// pulled before a timeout (if any).
func countBatches(ctx context.Context, cur *mongo.Cursor, batchSize int) (int, error) {
	count := 0
	batchCount := 0
	for cur.Next(ctx) {
		var res bson.D
		if err := cur.Decode(&res); err != nil {
			return 0, nil
		}

		count++

		if count%batchSize == 0 {
			batchCount++
		}
	}

	return batchCount, nil
}
