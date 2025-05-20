package event_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/event"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/mongo/writeconcern"
)

func TestCommandMonitor(t *testing.T) {
	monitor := &event.CommandMonitor{
		Started: func(_ context.Context, evt *event.CommandStartedEvent) {
			fmt.Println(evt.Command)
		},
	}

	client, err := mongo.Connect(options.Client().SetMonitor(monitor).SetWriteConcern(writeconcern.Majority()))
	require.NoError(t, err)

	defer client.Disconnect(context.Background())

	coll := client.Database("test").Collection("events")
	coll.InsertOne(context.Background(), bson.D{{Key: "name", Value: "test"}})
}
