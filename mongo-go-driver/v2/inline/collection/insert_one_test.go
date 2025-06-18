package collection_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func TestInsertOne(t *testing.T) {
	const uri = "mongodb://127.0.0.1:28017/?directConnection=true"

	clientOpts := options.Client().ApplyURI(uri)
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

	_, err = coll.InsertOne(context.Background(), map[string]interface{}{"name": "Alice", "age": 30})
	require.NoError(t, err)
}
