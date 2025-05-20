package collection

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/event"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Test how the global client-level timeout effects change stream.
func TestChangeStreamClientTimeout(t *testing.T) {
	m := &event.CommandMonitor{
		Started: func(_ context.Context, cse *event.CommandStartedEvent) {
			if cse.CommandName == "aggregate" {
				t.Log("aggregate: ", cse.Command)
			}
			if cse.CommandName == "getMore" {
				t.Log("getMore: ", cse.Command)
			}
		},
	}

	opts := options.Client().SetTimeout(1 * time.Second).SetMonitor(m)

	client, err := mongo.Connect(context.Background(), opts)
	require.NoError(t, err)

	defer func() {
		err := client.Disconnect(context.Background())
		require.NoError(t, err)
	}()

	db := client.Database("testdb")
	coll := db.Collection("coll")

	_ = coll.Drop(context.Background())

	changeStream, err := coll.Watch(context.Background(), mongo.Pipeline{})
	require.NoError(t, err)

	go func() {
		//ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		//defer cancel()

		ctx := context.Background()

		start := time.Now()
		for changeStream.Next(ctx) {
			var event bson.M
			err := changeStream.Decode(&event)
			require.NoError(t, err)
		}

		fmt.Println(time.Since(start))
	}()

	time.Sleep(2 * time.Second)

	_, err = coll.InsertOne(context.Background(), bson.D{{Key: "foo", Value: "bar"}})
	require.NoError(t, err)
}

func TestChangeStreamMaxAwaitTimeMS(t *testing.T) {
	// Command monitor to log aggregate and getMore commands.
	m := &event.CommandMonitor{
		Started: func(_ context.Context, cse *event.CommandStartedEvent) {
			if cse.CommandName == "aggregate" {
				t.Log("aggregate:", cse.Command)
			}
			if cse.CommandName == "getMore" {
				t.Log("getMore:", cse.Command)
			}
		},
	}

	// Configure the client with the command monitor.
	clientOpts := options.Client().
		SetMonitor(m)

	// Connect to the MongoDB server
	client, err := mongo.Connect(context.Background(), clientOpts)
	require.NoError(t, err)

	defer func() {
		err := client.Disconnect(context.Background())
		require.NoError(t, err)
	}()

	// Select the database and collection
	db := client.Database("testdb")
	coll := db.Collection("coll")

	// Drop the collection if it exists to start fresh
	_ = coll.Drop(context.Background())

	// Options for the change stream with maxAwaitTimeMS set
	changeStreamOpts := options.ChangeStream().SetMaxAwaitTime(5 * time.Second)

	// Start a change stream on the collection with maxAwaitTimeMS
	changeStream, err := coll.Watch(context.Background(), mongo.Pipeline{}, changeStreamOpts)
	require.NoError(t, err)

	// Run the change stream in a separate goroutine
	go func() {
		ctx := context.Background()
		start := time.Now()
		for changeStream.Next(ctx) {
			var event bson.M
			err := changeStream.Decode(&event)
			require.NoError(t, err)
			t.Logf("Change detected: %v", event)
		}
		fmt.Println("Duration for changes:", time.Since(start))
	}()

	// Wait for a bit before triggering a change
	time.Sleep(2 * time.Second)

	// Insert a document to create a change
	_, err = coll.InsertOne(context.Background(), bson.D{{Key: "foo", Value: "bar"}})
	require.NoError(t, err)

	_, err = coll.InsertOne(context.Background(), bson.D{{Key: "foo", Value: "bar1"}})
	require.NoError(t, err)

	_, err = coll.InsertOne(context.Background(), bson.D{{Key: "foo", Value: "bar2"}})
	require.NoError(t, err)

	time.Sleep(2 * time.Second)
}
