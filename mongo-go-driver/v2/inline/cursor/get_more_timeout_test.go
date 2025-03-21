package cursor

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func TestGetMoreTimeout(t *testing.T) {
	clientOpts := options.Client().SetMonitor(newCommandMonitor())

	client, err := mongo.Connect(clientOpts)
	require.NoError(t, err)

	defer func() {
		err := client.Disconnect(context.Background())
		require.NoError(t, err)
	}()

	db := client.Database("testdb")
	coll := db.Collection("coll")

	coll.Drop(context.Background())

	_, err = coll.InsertOne(context.Background(), bson.D{{"msg", "hello"}})
	require.NoError(t, err)

	_, err = coll.InsertOne(context.Background(), bson.D{{"msg", "hello"}})
	require.NoError(t, err)

	opts := options.Find().SetBatchSize(1)

	cur, err := coll.Find(context.Background(), bson.D{}, opts)
	require.NoError(t, err)

	cur.Next(context.Background())

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	cur.Next(ctx)
	cancel()

	fmt.Println(cur.Err())
}
