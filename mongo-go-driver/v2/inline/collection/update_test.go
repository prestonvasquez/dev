package collection_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

func TestUpdateImmutableIDField(t *testing.T) {
	client, err := mongo.Connect()
	require.NoError(t, err)

	defer client.Disconnect(context.Background())

	coll := client.Database("testdb").Collection("coll")

	err = coll.Drop(context.Background())
	require.NoError(t, err)

	_, err = coll.InsertOne(context.Background(), bson.D{
		{Key: "_id", Value: "abc123"},
		{Key: "name", Value: "original"},
	})
	require.NoError(t, err)

	filter := bson.D{{Key: "_id", Value: "abc123"}}
	update := bson.D{
		{Key: "_id", Value: "xyc789"},
		{Key: "name", Value: "replacement"},
	}

	_, err = coll.ReplaceOne(context.Background(), filter, update)
	fmt.Println(err)
}
