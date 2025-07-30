package collection

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/v2/event"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func TestFind(t *testing.T) {
	monitor := &event.CommandMonitor{
		Started: func(ctx context.Context, evt *event.CommandStartedEvent) {
			if evt.CommandName == "find" {
				t.Logf("CommandStartedEvent: %s, Command: %s", evt.CommandName, evt.Command)
			}
		},
	}

	clientOpts := options.Client().SetMonitor(monitor).SetTimeout(0)

	client, err := mongo.Connect(clientOpts)
	require.NoError(t, err)

	defer func() {
		err := client.Disconnect(context.Background())
		require.NoError(t, err)
	}()

	coll := client.Database("test").Collection("myCollection")
	defer func() {
		coll.Drop(context.Background())
	}()

	// Insert some documents to ensure the collection exists
	for i := 0; i < 5; i++ {
		coll.InsertOne(context.Background(), map[string]interface{}{"name": "Alice", "age": i + 20})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10)
	defer cancel()

	fintOpts := options.Find().SetBatchSize(0) // essentially an instant operation

	type f struct {
		Name   string
		Rating int
	}

	cur, err := coll.Find(ctx, f{Name: "meep"}, fintOpts)
	require.NoError(t, err)

	cur.SetBatchSize(1000) // Update

	for cur.Next(ctx) { // Apply the timeout at the iteration level
	}
}
