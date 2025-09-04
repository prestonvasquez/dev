package database_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/event"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func TestRunCommand_FailPoint(t *testing.T) {
	commandMonitor := &event.CommandMonitor{
		//Started: func(ctx context.Context, evt *event.CommandStartedEvent) {
		//	fmt.Println("Command started:", evt.CommandName, evt.Command)
		//},
		//Succeeded: func(ctx context.Context, evt *event.CommandSucceededEvent) {
		//	fmt.Println("Command succeeded:", evt.CommandName, evt.Reply)
		//},
		//Failed: func(ctx context.Context, evt *event.CommandFailedEvent) {
		//	fmt.Println("Command failed:", evt.CommandName, evt.Failure)
		//},
	}

	clientOpts := options.Client().
		ApplyURI("mongodb://localhost:27017").
		SetAppName("fp-demo").
		SetRetryReads(false). // important so the driver doesn’t retry the failed read
		SetMonitor(commandMonitor)

	client, err := mongo.Connect(clientOpts)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = client.Disconnect(context.Background()) })

	client.Database("test").Collection("testcoll").Drop(context.Background())

	// Create a capped collection
	err = client.Database("test").CreateCollection(context.Background(), "testcoll",
		options.CreateCollection().SetCapped(true).SetSizeInBytes(1024*64))

	require.NoError(t, err, "error creating capped collection")

	coll := client.Database("test").Collection("testcoll")

	// Insert seed data.
	docs := make([]interface{}, 100)
	for i := range docs {
		docs[i] = map[string]interface{}{"x": i}
	}

	_, err = coll.InsertMany(context.Background(), docs)
	require.NoError(t, err)

	ctx := context.Background()

	admin := client.Database("admin")

	// Fail exactly one find from this app by closing the connection
	if err := admin.RunCommand(ctx, bson.D{
		{"configureFailPoint", "failCommand"},
		{"mode", bson.D{{"times", 1}}},
		{"data", bson.D{
			{"failCommands", bson.A{"getMore"}},
			{"appName", "fp-demo"}, // restrict to this client
			//{"closeConnection", true}, // cause an observable failure
			{"blockConnection", true},
			{"blockTimeMS", 10_000}, // 10 seconds
		}},
	}).Err(); err != nil {
		t.Fatalf("configure failpoint failed: %v", err)
	}
	t.Cleanup(func() {
		_ = admin.RunCommand(context.Background(), bson.D{
			{"configureFailPoint", "failCommand"},
			{"mode", "off"},
		}).Err()
	})

	// Issue find via RunCommand (not the Find helper)
	db := client.Database("test")
	cur, err := db.RunCommandCursor(ctx, bson.D{
		{"find", "testcoll"},
		{"filter", bson.D{{"x", 6}}},
		{"tailable", true},
		{"awaitData", true},
		{"batchSize", int32(1)},
	})
	require.NoError(t, err, "error running command cursor")

	defer func() {
		err := cur.Close(ctx)
		assert.NoError(t, err, "error closing cursor")
	}()

	assert.True(t, cur.Next(context.Background()))
	assert.False(t, cur.Next(context.Background()))
	require.Error(t, cur.Err(), "expected error from cursor.Next")
}
