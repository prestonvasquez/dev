package cursor

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

func TestRunCommandCursor_TailableAwaitData(t *testing.T) {
	client, err := mongo.Connect()
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	defer func() { assert.NoError(t, client.Disconnect(ctx)) }()

	// Specify the database and collection
	db := client.Database("exampleDB")
	collection := db.Collection("exampleCappedCollection")
	collection.Drop(context.Background())
	defer func() { assert.NoError(t, collection.Drop(context.Background())) }()

	err = setupCappedCollection(context.Background(), collection)
	require.NoError(t, err)

	seedCollection(context.Background(), t, collection, 5)
}

func seedCollection(ctx context.Context, t *testing.T, coll *mongo.Collection, numDocs int) {
	t.Helper()

	docs := make([]interface{}, numDocs)
	for i := 0; i < numDocs; i++ {
		docs[i] = bson.D{{"x", i}}
	}

	_, err := coll.InsertMany(ctx, docs)
	require.NoError(t, err)
}
