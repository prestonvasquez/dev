package index

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func TestCreateVectorSearchIndex(t *testing.T) {
	// First create a client and ensure that a vector search index that supports
	// OpenAI's embedding model exists on the example collection.
	uri := os.Getenv("MONGODB_URI")
	if uri == "" {
		uri = "mongodb://localhost:27017"
	}

	client, err := mongo.Connect(options.Client().ApplyURI(uri))
	require.NoError(t, err)

	defer func() {
		err := client.Disconnect(context.Background())
		require.NoError(t, err)
	}()

	err = client.Database("db").CreateCollection(context.Background(), "vstore")
	require.NoError(t, err)

	//coll := client.Database("db").Collection("vstore")

	//fields := []vectorField{
	//	{
	//		Type:          "vector",
	//		Path:          "plot_embedding", // Default path
	//		NumDimensions: 1536,
	//		Similarity:    "euclidean",
	//	},
	//	{
	//		Type: "filter",
	//		Path: "metadata.area",
	//	},
	//	{
	//		Type: "filter",
	//		Path: "metadata.population",
	//	},
	//}

	//_, err = createVectorSearchIndex(context.Background(), coll, "vector_index_dotProduct_1536", fields...)
	//require.NoError(t, err)
}
