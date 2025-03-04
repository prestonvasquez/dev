package cursor

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func TestTailableAwaitDataMultipleGetMore(t *testing.T) {
	clientOpts := options.Client().SetMonitor(newCommandMonitor())

	client, err := mongo.Connect(clientOpts)
	require.NoError(t, err)

	defer func() {
		err := client.Disconnect(context.Background())
		require.NoError(t, err)
	}()

	db := client.Database("testdb")
	coll := db.Collection("cappedColl")

	coll.Drop(context.Background())

	err = setupCappedCollection(context.Background(), coll)
	require.NoError(t, err)

	_, err = coll.InsertOne(context.Background(), bson.D{{"msg", "hello"}})
	require.NoError(t, err)

	_, err = coll.InsertOne(context.Background(), bson.D{{"msg", "hello"}})
	require.NoError(t, err)

	opts := options.Find().SetBatchSize(1).SetCursorType(options.TailableAwait).
		SetMaxAwaitTime(50 * time.Millisecond)

	cur, err := coll.Find(context.Background(), bson.D{}, opts)
	require.NoError(t, err)

	defer cur.Close(context.Background())

	cur.Next(context.Background())

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Millisecond)
	defer cancel()

	cur.Next(ctx)
}
