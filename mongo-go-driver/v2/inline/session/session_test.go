package session_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/event"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/mongo/readconcern"
	"go.mongodb.org/mongo-driver/v2/mongo/readpref"
	"go.mongodb.org/mongo-driver/v2/mongo/writeconcern"
)

func newCommandMonitor() *event.CommandMonitor {
	return &event.CommandMonitor{
		Started: func(ctx context.Context, evt *event.CommandStartedEvent) {
			fmt.Printf("Command started: %s\n", evt.Command)
		},
		//Succeeded: func(ctx context.Context, evt *event.CommandSucceededEvent) {
		//	fmt.Printf("Command succeeded: %s\n", evt.Command)
		//},
		//Failed: func(ctx context.Context, evt *event.CommandFailedEvent) {
		//	fmt.Printf("Command failed: %s\n", evt.Failure)
		//},
	}
}

func TestSameSessionMultipleCursor(t *testing.T) {
	client, err := mongo.Connect()
	require.NoError(t, err)

	defer client.Disconnect(context.Background())

	sess, err := client.StartSession()
	require.NoError(t, err)

	coll := client.Database("testdb").Collection("tstcoll")
	_ = coll.Drop(context.Background())

	ctx := mongo.NewSessionContext(context.Background(), sess)

	// Seed some data
	for i := 0; i < 10; i++ {
		coll.InsertOne(ctx, bson.D{{"status", "active"}})
		coll.InsertOne(ctx, bson.D{{"status", "inactive"}})
	}

	cur1, err := coll.Find(ctx, bson.D{{"status", "active"}})
	require.NoError(t, err)

	cur2, err := coll.Find(ctx, bson.D{{"status", "inactive"}})
	require.NoError(t, err)

	allActive := []bson.D{}
	require.NoError(t, cur1.All(ctx, &allActive))
	assert.Len(t, allActive, 10)

	allInactive := []bson.D{}
	require.NoError(t, cur2.All(ctx, &allInactive))
	assert.Len(t, allInactive, 10)
}

// (InvalidOptions) writeConcern is not allowed within a multi-statement transaction
func TestInvalidWriteConcernInTransaction(t *testing.T) {
	clientOpts := options.Client().
		SetWriteConcern(writeconcern.Majority()).
		SetReadConcern(readconcern.Majority()).
		SetReadPreference(readpref.PrimaryPreferred()).
		SetMonitor(newCommandMonitor())

	client, err := mongo.Connect(clientOpts)
	require.NoError(t, err)

	defer client.Disconnect(context.Background())

	sessionOpts := options.Session().SetCausalConsistency(true)
	session, err := client.StartSession(sessionOpts)
	require.NoError(t, err)

	defer session.EndSession(context.Background())

	txnOpts := options.Transaction().SetReadPreference(readpref.Primary())
	fmt.Println("Starting transaction with invalid write concern")
	_, err = session.WithTransaction(context.Background(), func(sc context.Context) (interface{}, error) {

		coll := client.Database("your_database_name").Collection("coll")
		res, err := coll.InsertOne(sc, bson.D{{"_id", 0}})
		fmt.Println("done")
		return res, err
		// Create the insert command
		//command := bson.D{
		//	{Key: "insert", Value: "t"},
		//	{Key: "documents", Value: bson.A{bson.D{}}},
		//	{Key: "writeConcern", Value: bson.D{{Key: "w", Value: 1}}},
		//}

		//var result bson.M
		//// Run the command in the session context
		//err := client.Database("your_database_name").RunCommand(sc, command).Decode(&result)
		//if err != nil {
		//	return nil, fmt.Errorf("failed to run command: %w", err)
		//}

		//fmt.Println("Command result:", result)
		//return result, nil
	}, txnOpts)

	require.NoError(t, err)
}

// TestCursorNotFoundPinningBug reproduces a potential driver bug where getMore commands
// are not properly pinned to the same server as the original find command.
//
// This test verifies that:
// 1. The initial find command goes to a specific server (with secondary read preference)
// 2. Subsequent getMore commands go to the SAME server
// 3. If they don't, a CursorNotFound error will occur
//
// To reproduce the bug, this test:
// - Uses a replica set with secondaries
// - Sets read preference to secondary
// - Creates a cursor with batch size to force getMore calls
// - Monitors command events to verify server pinning
func TestCursorNotFoundPinningBug(t *testing.T) {
	type commandEvent struct {
		commandName string
		cursorID    int64
		serverAddr  string
		connID      string
	}

	events := []commandEvent{}

	// Command monitor to track server pinning
	cmdMonitor := &event.CommandMonitor{
		Started: func(_ context.Context, evt *event.CommandStartedEvent) {
			if evt.CommandName == "find" || evt.CommandName == "getMore" {
				var cursorID int64
				if evt.CommandName == "getMore" {
					if val, err := evt.Command.LookupErr("getMore"); err == nil {
						cursorID = val.Int64()
					}
				}

				events = append(events, commandEvent{
					commandName: evt.CommandName,
					cursorID:    cursorID,
					serverAddr:  evt.ServerConnectionID.Address().String(),
					connID:      evt.ConnectionID,
				})

				t.Logf("%s command to server=%s, connectionID=%s, cursorID=%d",
					evt.CommandName,
					evt.ServerConnectionID.Address(),
					evt.ConnectionID,
					cursorID)
			}
		},
		Failed: func(_ context.Context, evt *event.CommandFailedEvent) {
			if evt.CommandName == "getMore" {
				t.Logf("getMore FAILED: %v, server=%s, connectionID=%s",
					evt.Failure,
					evt.ServerConnectionID.Address(),
					evt.ConnectionID)
			}
		},
	}

	// Configure client with read preference to secondary
	clientOpts := options.Client().
		SetReadPreference(readpref.Secondary()).
		SetMonitor(cmdMonitor)

	client, err := mongo.Connect(clientOpts)
	require.NoError(t, err)
	defer client.Disconnect(context.Background())

	// Use explicit session to ensure cursor pinning
	sess, err := client.StartSession()
	require.NoError(t, err)
	defer sess.EndSession(context.Background())

	coll := client.Database("testdb").Collection("cursor_test")
	ctx := mongo.NewSessionContext(context.Background(), sess)

	// Drop and seed with enough data to force multiple getMore calls
	_ = coll.Drop(context.Background())

	docs := make([]interface{}, 100)
	for i := 0; i < 100; i++ {
		docs[i] = bson.D{{"index", i}, {"data", fmt.Sprintf("document-%d", i)}}
	}
	_, err = coll.InsertMany(context.Background(), docs)
	require.NoError(t, err)

	// Find with small batch size to force getMore calls
	findOpts := options.Find().SetBatchSize(10)
	cursor, err := coll.Find(ctx, bson.D{}, findOpts)
	require.NoError(t, err)
	defer cursor.Close(ctx)

	// Iterate through all documents, triggering multiple getMore commands
	count := 0
	for cursor.Next(ctx) {
		var result bson.D
		err := cursor.Decode(&result)
		require.NoError(t, err)
		count++
	}

	require.NoError(t, cursor.Err())
	assert.Equal(t, 100, count, "Should have retrieved all 100 documents")

	// Verify server pinning: all commands should go to the same server
	if len(events) > 0 {
		firstServer := events[0].serverAddr
		t.Logf("\nVerifying cursor pinning:")
		t.Logf("Initial find went to: %s", firstServer)

		for i, evt := range events {
			t.Logf("  [%d] %s -> server=%s", i, evt.commandName, evt.serverAddr)
			assert.Equal(t, firstServer, evt.serverAddr,
				"Command %s should be pinned to the same server as find. "+
					"This indicates a cursor pinning bug in the driver.",
				evt.commandName)
		}

		// Verify we actually had getMore commands (otherwise test is not useful)
		getMoreCount := 0
		for _, evt := range events {
			if evt.commandName == "getMore" {
				getMoreCount++
			}
		}
		t.Logf("\nTotal getMore commands: %d", getMoreCount)
		assert.Greater(t, getMoreCount, 0, "Test should trigger at least one getMore command")
	}
}
