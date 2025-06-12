package collection_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/v2/event"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/mongo/readpref"
)

func TestCloneReadPreference(t *testing.T) {
	// Create a command monitor.
	monitor := &event.CommandMonitor{
		Started: func(ctx context.Context, evt *event.CommandStartedEvent) {
			cmd := evt.Command
			fmt.Println(cmd)
		},
	}

	client, err := mongo.Connect(options.Client().SetMonitor(monitor))
	require.NoError(t, err)

	defer client.Disconnect(context.Background())

	collOpts := options.Collection().SetReadPreference(readpref.SecondaryPreferred())
	coll := client.Database("testdb").Collection("test", collOpts) // Expected seccondary preferred read preference

	coll.FindOne(context.Background(), map[string]string{"name": "test"})

	cloneCollOpts := options.Collection().SetReadPreference(readpref.PrimaryPreferred())
	clonedColl := coll.Clone(cloneCollOpts) // Expected primary preferred read preference

	clonedColl.FindOne(context.Background(), map[string]string{"name": "test"})
}
