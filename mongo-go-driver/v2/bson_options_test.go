package v2

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func TestBsonOptions_ObjectIDAsHexString(t *testing.T) {
	t.Run("unset", func(t *testing.T) {
		client, err := mongo.Connect()
		require.NoError(t, err)

		defer func() { _ = client.Disconnect(context.Background()) }()

		coll := client.Database("db").Collection("coll")
		defer func() { coll.Drop(context.Background()) }()

		// Insert some test data to query.
		_, err = coll.InsertOne(context.Background(), bson.D{{"x", 1}})
		require.NoError(t, err)

		// Query the data and try to get the id as a string.
		res := coll.FindOne(context.Background(), bson.D{{"x", 1}})
		require.NoError(t, res.Err())

		type data struct {
			ID string `bson:"_id"`
		}

		d := data{}
		err = res.Decode(&d)
		assert.Error(t, err)
	})

	t.Run("set", func(t *testing.T) {
		opts := options.Client().SetBSONOptions(&options.BSONOptions{ObjectIDAsHexString: true})
		client, err := mongo.Connect(opts)
		require.NoError(t, err)

		defer func() { _ = client.Disconnect(context.Background()) }()

		coll := client.Database("db").Collection("coll")
		defer func() { coll.Drop(context.Background()) }()

		// Insert some test data to query.
		_, err = coll.InsertOne(context.Background(), bson.D{{"x", 1}})
		require.NoError(t, err)

		// Query the data and try to get the id as a string.
		res := coll.FindOne(context.Background(), bson.D{{"x", 1}})
		require.NoError(t, res.Err())

		type data struct {
			ID string `bson:"_id"`
		}

		d := data{}
		err = res.Decode(&d)
		assert.NoError(t, err)

		t.Log("Data: ", d)
	})
}
