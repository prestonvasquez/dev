package v1

import (
	"context"
	"os"
	"testing"

	"github.com/prestonvasquez/dev/mongo-go-driver/v1/util/indexutil"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func TestVectorStore(t *testing.T) {
	uri := os.Getenv("MONGODB_URI")
	assert.NotEmpty(t, uri, "MONGOD_URI must be set as an Atlas cluster")

	opts := options.Client().ApplyURI(uri)

	client, err := mongo.Connect(context.Background(), opts)
	assert.NoError(t, err, "failed to connect")

	defer client.Disconnect(context.Background())

	const dbName = "db"
	const collName = "vstore"

	// Create the vectorstore collection
	err = client.Database(dbName).CreateCollection(context.Background(), collName)
	assert.NoError(t, err, "failed to create vstore collection")

	coll := client.Database(dbName).Collection(collName)
	defer coll.Drop(context.Background())

	// Create the index
	field := indexutil.VectorField{
		Type:          "vector",
		Path:          "plot_embedding",
		NumDimensions: 1536,
		Similarity:    indexutil.VectorSimilarityDotProduct,
	}

	_, err = indexutil.CreateVectorSearch(context.Background(), coll, field)
	assert.NoError(t, err, "failed to create vector search index")
}
