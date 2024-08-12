package indexutil

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// CreateN will create n arbitrary indexes.
func CreateN(ctx context.Context, coll *mongo.Collection, n int) error {
	models := make([]mongo.IndexModel, n)
	for i := 0; i < n; i++ {
		name := fmt.Sprintf("name_%v", i)

		models[i] = mongo.IndexModel{Keys: bson.D{{name, 1}}}
	}

	_, err := coll.Indexes().CreateMany(ctx, models)

	return err
}

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

// CreateVectorSearch will create a vector search index.
func CreateVectorSearch(ctx context.Context, coll *mongo.Collection, fields ...VectorField) (string, error) {
	if coll == nil {
		return "", fmt.Errorf("coll or idx must not be nil")
	}

	if len(fields) == 0 {
		return "", nil
	}

	def := struct {
		Mappings struct {
			Dynamic bool `bson:"dynamic"`
		} `bson:"mappings"`
		Fields []VectorField `bson:"fields"`
	}{
		Fields: append([]VectorField{}, fields...),
	}

	view := coll.SearchIndexes()

	searchName, err := view.CreateOne(ctx, mongo.SearchIndexModel{Definition: def})
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
