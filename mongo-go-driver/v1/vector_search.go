package v1

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"math"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const OpenAITextEmbedding3SmallSize = 1536

// OpenAITextEmbedding3SmallSize is in 1536-space

// VectorSimilarity defines which function to use to search for top K-nearest
// neighbors.
type VectorSimilarity string

const (
	// VectorSimilarityCosine measures similarity based on the angle between
	// vectors. This value allows you to measure similarity that isn't scaled by
	// magnitude. You can't use zero magnitude vectors with cosine. To measure
	// cosine similarity, we recommend that you normalize your vectors and use
	// dotProduct instead.
	VectorSimilarityCosine = "cosine"

	// VectorSimilarityDotProduct measures similar to cosine, but takes into
	// account the magnitude of the vector. This value allows you to efficiently
	// measure similarity based on both angle and magnitude. To use dotProduct,
	// you must normalize the vector to unit length at index-time and query-time.
	VectorSimilarityDotProduct = "dotProduct"

	// VectorSimilarityEuclidean measures the distance between ends of vectors.
	// This value allows you to measure similarity based on varying dimensions.
	VectorSimilarityEuclidean = "euclidean"
)

// VectorField defines the fields of an index used for vector search.
type VectorField struct {
	Type          string           `bson:"type"`
	Path          string           `bson:"path"`
	NumDimensions int              `bson:"numDimensions"`
	Similarity    VectorSimilarity `bson:"similarity"`
}

// CreateMockEmbedding will create a float32 vector with random elements of size
// n for mocking an embedding.
func CreateMockEmbedding(n int) ([]float32, error) {
	f32s := make([]float32, n)

	for i := range f32s {
		var b [4]byte

		_, err := rand.Read(b[:])
		if err != nil {
			return nil, fmt.Errorf("failed to read random bytes: %w", err)
		}

		f32s[i] = float32(binary.LittleEndian.Uint32(b[:])) / math.MaxUint32
	}

	return f32s, nil
}

// CreateVectorSearchIndex will create a vector search index on the "db.vstore"
// collection named "vector_index" with the provided field. This function blocks
// until the index has been created.
func CreateVectorSearchIndex(
	ctx context.Context,
	coll *mongo.Collection,
	idxName string,
	field VectorField,
) (string, error) {
	def := struct {
		Fields []VectorField `bson:"fields"`
	}{
		Fields: []VectorField{field},
	}

	view := coll.SearchIndexes()

	siOpts := options.SearchIndexes().SetName(idxName).SetType("vectorSearch")
	searchName, err := view.CreateOne(ctx, mongo.SearchIndexModel{Definition: def, Options: siOpts})
	if err != nil {
		return "", fmt.Errorf("failed to create the search index: %w", err)
	}

	// Await the creation of the index.
	var doc bson.Raw
	for doc == nil {
		cursor, err := view.List(ctx, options.SearchIndexes().SetName(searchName))
		if err != nil {
			return "", fmt.Errorf("failed to list search indexes: %w", err)
		}

		if !cursor.Next(ctx) {
			break
		}

		name := cursor.Current.Lookup("name").StringValue()
		queryable := cursor.Current.Lookup("queryable").Boolean()
		if name == searchName && queryable {
			doc = cursor.Current
		} else {
			time.Sleep(5 * time.Second)
		}
	}

	if doc == nil {
		return "", fmt.Errorf("index not found or empty")
	}

	return searchName, nil
}

// DropVectorSearchIndex will attempt to drop the search index by name, awaiting
// that it has been dropped. This function blocks until the index has been
// dropped.
func DropVectorSearchIndex(ctx context.Context, coll *mongo.Collection, idxName string) error {
	if coll == nil {
		return fmt.Errorf("collection must not be nil")
	}

	view := coll.SearchIndexes()

	if err := view.DropOne(ctx, idxName); err != nil {
		return fmt.Errorf("failed to drop index: %w", err)
	}

	// Await the drop of the index.
	for {
		cursor, err := view.List(ctx, options.SearchIndexes().SetName(idxName))
		if err != nil {
			return fmt.Errorf("failed to list search indexes: %w", err)
		}

		if !cursor.Next(ctx) {
			break
		} else {
			time.Sleep(5 * time.Second)
		}
	}

	return nil
}

// SearchVectors will perform a similarity search on the vector space. A basic
// similarity search should look something like this:
//
// vector_store.similarity_search(query="thud",k=1)
//
// The user will provide a query. We need to get the embedding
// and then add that to the aggregation pipeline's stage.
//
// In practice, we would use the embeddings.Embedder to do this:
//
// openAIEmbedder, _ := embeddings.NewEmbedder(llm)
// vector, _ := openAIEmbedder.EmbedQuery(ctx, query)
func SearchVectors(ctx context.Context, coll *mongo.Collection, vectorSearchStage bson.D) ([]bson.D, error) {
	// Create a vector search pipeline
	pipeline := mongo.Pipeline{
		bson.D{
			{"$vectorSearch", vectorSearchStage},
		},
		bson.D{
			{"$set", bson.D{{"score", bson.D{{"$meta", "vectorSearchScore"}}}}},
		},
	}

	cur, err := coll.Aggregate(context.Background(), pipeline)
	if err != nil {
		return nil, fmt.Errorf("failed to aggregate: %v", err)
	}

	var found []bson.D
	for cur.Next(context.Background()) {
		var d bson.D
		err := cur.Decode(&d)
		if err != nil {
			return nil, fmt.Errorf("failed to decode: %v", err)
		}

		found = append(found, d)
	}

	return found, nil
}
