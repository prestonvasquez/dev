package cursor

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/event"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func TestFindWithMaxAwaitTimeValidation(t *testing.T) {
	cmdMonitor := &event.CommandMonitor{
		Started: func(ctx context.Context, evt *event.CommandStartedEvent) {
			if evt.CommandName == "getMore" {
				t.Logf("CommandStartedEvent: %s, Command: %s", evt.CommandName, evt.Command)
			}
		},
	}

	client, err := mongo.Connect(options.Client().SetMonitor(cmdMonitor).SetTimeout(20 * time.Second))
	require.NoError(t, err, "failed to connect to server")

	t.Cleanup(func() {
		_ = client.Disconnect(context.Background())
	})

	db := client.Database("exampleDB")
	collName := "exampleCappedCollection"

	// Drop the collection if it exists
	_ = db.Collection(collName).Drop(context.Background())

	// Create a capped collection.
	collOpts := options.CreateCollection().SetCapped(true).SetSizeInBytes(1024 * 1024) // 1 MB capped collection

	err = db.CreateCollection(context.Background(), collName, collOpts)
	require.NoError(t, err, "failed to create capped collection")

	coll := db.Collection(collName)

	// Insert some documents to ensure the collection exists
	_, err = coll.InsertMany(context.Background(), []interface{}{
		bson.D{{Key: "name", Value: "Alice"}},
		bson.D{{Key: "name", Value: "Bob"}},
		bson.D{{Key: "name", Value: "Charlie"}},
	})
	require.NoError(t, err, "failed to insert documents")

	opts := options.Find().SetMaxAwaitTime(10 * time.Second).SetBatchSize(3).SetCursorType(options.TailableAwait)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	res, err := coll.Find(ctx, bson.D{}, opts)
	if err != nil {
		t.Logf("error running Find with MaxAwaitTime: %v", err)
	}

	fmt.Println("Cursor created with MaxAwaitTime set to 10 seconds")

	//assert.Error(t, err, "expected validation error when using MaxAwaitTime with a non-capped collection")

	// What happens if we use the decoder?
	for res.Next(ctx) {
		fmt.Println("Found document:", res.Current)
	}

	if err := res.Err(); err != nil {
		t.Errorf("cursor error: %v", err)
	}
}

func TestClientLevelTimeoutForCursors(t *testing.T) {
	client, err := mongo.Connect(options.Client().SetTimeout(500 * time.Millisecond))
	require.NoError(t, err, "failed to connect to server")

	t.Cleanup(func() {
		_ = client.Disconnect(context.Background())
	})

	db := client.Database("exampleDB")
	collName := "exampleCappedCollection"

	// Drop the collection if it exists
	_ = db.Collection(collName).Drop(context.Background())

	// Create a capped collection.
	collOpts := options.CreateCollection().SetCapped(true).SetSizeInBytes(1024 * 1024) // 1 MB capped collection

	err = db.CreateCollection(context.Background(), collName, collOpts)
	require.NoError(t, err, "failed to create capped collection")

	coll := db.Collection(collName)

	// Insert some documents to ensure the collection exists
	_, err = coll.InsertMany(context.Background(), []interface{}{
		bson.D{{Key: "name", Value: "Alice"}},
		bson.D{{Key: "name", Value: "Bob"}},
		bson.D{{Key: "name", Value: "Charlie"}},
	})
	require.NoError(t, err, "failed to insert documents")

	opts := options.Find().SetMaxAwaitTime(10 * time.Second).SetBatchSize(3).SetCursorType(options.TailableAwait)
	res, err := coll.Find(context.Background(), bson.D{}, opts)
	if err != nil {
		t.Logf("error running Find with MaxAwaitTime: %v", err)
	}

	fmt.Println("Cursor created with MaxAwaitTime set to 10 seconds")

	//assert.Error(t, err, "expected validation error when using MaxAwaitTime with a non-capped collection")

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// What happens if we use the decoder?
	for res.Next(ctx) {
		fmt.Println("Found document:", res.Current)
	}

	if err := res.Err(); err != nil {
		t.Errorf("cursor error: %v", err)
	}
}

// Non TailableAwait
func TestCursorLevelTimeouts(t *testing.T) {
	cmdMonitor := &event.CommandMonitor{
		Started: func(ctx context.Context, evt *event.CommandStartedEvent) {
			if evt.CommandName == "getMore" {
				t.Logf("CommandStartedEvent: %s, Command: %s", evt.CommandName, evt.Command)
			}
		},
	}

	client, err := mongo.Connect(options.Client().SetMonitor(cmdMonitor).SetTimeout(20 * time.Second))
	require.NoError(t, err, "failed to connect to server")

	t.Cleanup(func() {
		_ = client.Disconnect(context.Background())
	})

	db := client.Database("exampleDB")
	collName := "exampleCappedCollection"

	// Drop the collection if it exists
	_ = db.Collection(collName).Drop(context.Background())

	// Create a capped collection.
	collOpts := options.CreateCollection().SetCapped(true).SetSizeInBytes(1024 * 1024) // 1 MB capped collection

	err = db.CreateCollection(context.Background(), collName, collOpts)
	require.NoError(t, err, "failed to create capped collection")

	coll := db.Collection(collName)

	// Insert some documents to ensure the collection exists
	_, err = coll.InsertMany(context.Background(), []interface{}{
		bson.D{{Key: "name", Value: "Alice"}},
		bson.D{{Key: "name", Value: "Bob"}},
		bson.D{{Key: "name", Value: "Charlie"}},
	})
	require.NoError(t, err, "failed to insert documents")

	opts := options.Find().SetMaxAwaitTime(10 * time.Second).SetBatchSize(0)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	res, err := coll.Find(ctx, bson.D{}, opts)
	if err != nil {
		t.Logf("error running Find with MaxAwaitTime: %v", err)
	}

	fmt.Println("Cursor created with MaxAwaitTime set to 10 seconds")

	res.SetBatchSize(1) // Set batch size to 3
	curCtx, curCancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer curCancel()

	// What happens if we use the decoder?
	for res.Next(curCtx) {
		fmt.Println("Found document:", res.Current)
	}

	if err := res.Err(); err != nil {
		t.Errorf("cursor error: %v", err)
	}
}
