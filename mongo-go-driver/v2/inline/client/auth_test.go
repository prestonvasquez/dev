package client

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func TestScram(t *testing.T) {
	const uri = "mongodb://bob:pwd123@localhost:27017"
	opts := options.Client().ApplyURI(uri)

	client, err := mongo.Connect(opts)
	require.NoError(t, err, "failed to connect to MongoDB")

	defer func() {
		err := client.Disconnect(context.Background())
		require.NoError(t, err, "failed to disconnect from MongoDB")
	}()

	//require.NoError(t, client.Ping(context.Background(), nil), "failed to ping MongoDB")
	coll := client.Database("test").Collection("test")

	res := coll.FindOne(context.Background(), bson.M{"name": "test"})
	require.NoError(t, res.Err(), "failed to find document")
}
