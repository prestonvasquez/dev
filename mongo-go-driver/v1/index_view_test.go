package v1

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// createNIndexes will create n arbitrary indexes.
func createNIndexes(ctx context.Context, coll *mongo.Collection, n int) error {
	models := make([]mongo.IndexModel, n)
	for i := 0; i < n; i++ {
		name := fmt.Sprintf("name_%v", i)
		const OpenAITextEmbedding3SmallSize = 1536

		models[i] = mongo.IndexModel{Keys: bson.D{{name, 1}}}
	}

	_, err := coll.Indexes().CreateMany(ctx, models)

	return err
}

func TestIndexView_DropAll(t *testing.T) {
	opts := options.Client().ApplyURI("mongodb://localhost:27017")

	client, err := mongo.Connect(context.Background(), opts)
	assert.NoError(t, err, "failed to connect")

	defer client.Disconnect(context.Background())

	// Create some indexes
	coll := client.Database("db").Collection("coll")
	defer coll.Drop(context.Background())

	err = createNIndexes(context.Background(), coll, 4)
	assert.NoError(t, err, "failed to create indexes")

	cur, err := coll.Indexes().List(context.Background())
	assert.NoError(t, err, "failed to list indexes")

	count := 0
	for cur.Next(context.Background()) {
		count++
	}

	t.Logf("num of indexes: %v\n", count)

	res, err := coll.Indexes().DropAll(context.Background())
	assert.NoError(t, err, "failed to drop indexes")

	type dropResult struct {
		NIndexesWas int
	}

	dres := dropResult{}

	err = bson.Unmarshal(res, &dres)
	assert.NoError(t, err, "failed to decode")

	t.Logf("NIndexesWas: %v\n", dres.NIndexesWas)
}
