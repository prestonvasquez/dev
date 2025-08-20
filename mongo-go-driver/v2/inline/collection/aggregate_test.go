package collection_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/event"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func TestAggregate_SetMaxAwaitTime_SetBatchSize(t *testing.T) {
	commandMonitor := &event.CommandMonitor{
		Started: func(ctx context.Context, evt *event.CommandStartedEvent) {
			if evt.CommandName == "aggregate" {
				fmt.Println("aggregate: ", evt.Command)
			}
			if evt.CommandName == "getMore" {
				fmt.Println("getMore: ", evt.Command)
			}
		},
	}

	clientOpts := options.Client().SetTimeout(200 * time.Second).SetMonitor(commandMonitor)

	client, err := mongo.Connect(clientOpts)
	require.NoError(t, err)

	defer func() {
		err := client.Disconnect(nil)
		assert.NoError(t, err)
	}()

	coll := client.Database("test").Collection("aggregate_test")

	// Insert seed data.
	docs := make([]interface{}, 100)
	for i := range docs {
		docs[i] = map[string]interface{}{"x": i}
	}

	_, err = coll.InsertMany(context.Background(), docs)
	require.NoError(t, err)

	pipeline := mongo.Pipeline{{{"$match", bson.D{}}}}

	opts := options.Aggregate().SetMaxAwaitTime(100 * time.Millisecond).SetBatchSize(1)

	cur, err := coll.Aggregate(context.Background(), pipeline, opts)
	require.NoError(t, err)

	defer func() {
		err := cur.Close(context.Background())
		assert.NoError(t, err)
	}()

	cur.Next(context.Background())
	cur.Next(context.Background())

	fmt.Println(cur.Err())
}
