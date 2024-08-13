package v1

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/prestonvasquez/dev/mongo-go-driver/v1/util"
	"github.com/prestonvasquez/dev/mongo-go-driver/v1/util/indexutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// How to create a vector search index.
func TestSearchIndex_CreateOne_LangchaingoPOC(t *testing.T) {
	uri := os.Getenv("MONGODB_URI")
	require.NotEmpty(t, uri, "MONGOD_URI must be set as an Atlas cluster")

	opts := options.Client().ApplyURI(uri)

	client, err := mongo.Connect(context.Background(), opts)
	require.NoError(t, err, "failed to connect")

	defer client.Disconnect(context.Background())

	const dbName = "db"
	const collName = "vstore"
	const indexName = "vector_index"
	const path = "plot_embedding"

	// Create the vectorstore collection
	err = client.Database(dbName).CreateCollection(context.Background(), collName)
	assert.NoError(t, err, "failed to create vstore collection")

	coll := client.Database(dbName).Collection(collName)
	defer coll.Drop(context.Background())

	// Drop the vector search index before re-creating it for the test.
	_ = indexutil.DropVectorSearch(context.Background(), coll, indexName)

	// Create the index
	field := indexutil.VectorField{
		Type:          "vector",
		Path:          path,
		NumDimensions: util.OpenAITextEmbedding3SmallSize,
		Similarity:    indexutil.VectorSimilarityEuclidean,
	}
	_, err = indexutil.CreateVectorSearch(context.Background(), coll, indexName, field)
	require.NoError(t, err, "failed to create vector search index")

	t.Logf("index name: %+v\n", indexName)

	// Create some mock vectors to embed.
	validMockVector, err := util.CreateVector(util.OpenAITextEmbedding3SmallSize)
	assert.NoError(t, err, "failed to create valid mock vector")

	invalidMockVector, err := util.CreateVector(util.OpenAITextEmbedding3SmallSize - 1)
	assert.NoError(t, err, "failed to create invalid mock vector")

	// Insert a document with a mock embedding.
	t.Run("insert mock vector", func(t *testing.T) {
		doc := bson.D{
			{"text", "test"},
			{path, validMockVector},
		}

		_, err = coll.InsertOne(context.Background(), doc)
		assert.NoError(t, err, "failed to insert embedding")
	})

	// Insert a document with an invalid number of dimensions.
	t.Run("insert mock valid with invalid vector size", func(t *testing.T) {
		doc := bson.D{
			{"text", "test"},
			{path, invalidMockVector},
		}

		_, err = coll.InsertOne(context.Background(), doc)
		assert.NoError(t, err, "failed to insert embedding")
	})

	// vector_store.similarity_search(query="thud",k=1)
	// A basic similarity search should look something like this:
	t.Run("similarity search", func(t *testing.T) {
		// The user will provide a query. We need to get the embedding
		// and then add that to the aggregation pipeline's stage.
		//
		// In practice, we would use the embeddings.Embedder to do this:
		//
		// openAIEmbedder, _ := embeddings.NewEmbedder(llm)
		// vector, _ := openAIEmbedder.EmbedQuery(ctx, query)
		stage := bson.D{
			// Name of Atlas Vector Search Index tied to Collection
			{"index", indexName},

			// Field in Collection containing embedding vectors
			{"path", path},
			{"queryVector", validMockVector},

			// List of embedding vector
			{"numCandidates", 150},
			{"limit", 10},
		}

		// Create a vector search pipeline
		pipeline := mongo.Pipeline{
			bson.D{
				{"$vectorSearch", stage},
			},
		}

		// TODO: why do we have to sleep for this to work consistently?
		time.Sleep(1 * time.Second)

		cur, err := coll.Aggregate(context.Background(), pipeline)
		require.NoError(t, err, "failed to aggregate")

		var found []bson.D
		for cur.Next(context.Background()) {
			var d bson.D
			err := cur.Decode(&d)
			assert.NoError(t, err, "failed to decode")

			found = append(found, d)
		}

		// We inserted one valid vector.
		assert.Equal(t, 1, len(found))
	})
}
