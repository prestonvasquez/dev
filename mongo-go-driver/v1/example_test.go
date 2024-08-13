package v1

import (
	"context"
	"fmt"
	"log"
	"math"
	"os"
	"reflect"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const testIdxName = "test_vector_index"

// createTestSearchIndexColl returns creates and returns "vstore" collection
// for the "db" database.
func createTestSearchIndexColl(ctx context.Context) (*mongo.Collection, func(context.Context), error) {
	cleanup := func(context.Context) {}

	uri := os.Getenv("MONGODB_URI")
	if uri == "" {
		return nil, cleanup, fmt.Errorf("MONGOD_URI must be set as an Atlas cluster")
	}

	opts := options.Client().ApplyURI(uri)

	client, err := mongo.Connect(ctx, opts)
	if err != nil {
		return nil, cleanup, fmt.Errorf("failed to connect: %w", err)
	}

	cleanup = func(ctx context.Context) { client.Disconnect(ctx) }

	const (
		dbName   = "db"
		collName = "vstore"
	)

	// Create the vectorstore collection
	err = client.Database(dbName).CreateCollection(ctx, collName)
	if err != nil {
		return nil, cleanup, fmt.Errorf("failed to create collection: %w", err)
	}

	coll := client.Database(dbName).Collection(collName)

	return coll, cleanup, nil
}

// createTestSearchIndex will drop the search index on the collection, if it
// exists, then re-create it.
func createTestSearchIndex(ctx context.Context, coll *mongo.Collection) (string, error) {
	// Drop the vector search index before re-creating it for the test.
	_ = DropVectorSearchIndex(ctx, coll, testIdxName)

	field := VectorField{
		Type:          "vector",
		Path:          "plot_embedding",
		NumDimensions: OpenAITextEmbedding3SmallSize,
		Similarity:    VectorSimilarityEuclidean,
	}

	actualIdxName, err := CreateVectorSearchIndex(ctx, coll, testIdxName, field)
	if err != nil {
		return "", fmt.Errorf("failed to create vector search index: %v", err)
	}

	return actualIdxName, nil
}

func ExampleCreateVectorSearchIndex() {
	coll, teardown, err := createTestSearchIndexColl(context.Background())
	if err != nil {
		log.Fatalf("failed to create test search index: %v", err)
	}

	defer teardown(context.Background())

	actualIdxName, err := createTestSearchIndex(context.Background(), coll)
	if err != nil {
		log.Fatalf("failed to create test search index: %v", err)
	}

	fmt.Println(actualIdxName)
	// Output: vector_index
}

func ExampleSearchVectors() {
	coll, teardown, err := createTestSearchIndexColl(context.Background())
	if err != nil {
		log.Fatalf("failed to create test search index: %v", err)
	}

	defer teardown(context.Background())
	defer coll.Drop(context.Background())

	if _, err := createTestSearchIndex(context.Background(), coll); err != nil {
		log.Fatalf("failed to create test search index: %v", err)
	}

	// Add documents to the collection.
	fooVector, err := CreateMockEmbedding(OpenAITextEmbedding3SmallSize)
	if err != nil {
		log.Fatalf("failed to create valid mock vector: %v", err)
	}

	barVector, err := CreateMockEmbedding(OpenAITextEmbedding3SmallSize)
	if err != nil {
		log.Fatalf("failed to create valid mock vector: %v", err)
	}

	if reflect.DeepEqual(fooVector, barVector) {
		log.Fatal("foo and bar vectors are the same")
	}

	docs := []interface{}{
		bson.D{
			{"text", "foo"},
			{"plot_embedding", fooVector},
		},
		bson.D{
			{"text", "bar"},
			{"plot_embedding", barVector},
		},
	}

	if _, err := coll.InsertMany(context.Background(), docs); err != nil {
		log.Fatalf("failed to insert documents: %v", err)
	}

	// TODO: Why do we have to wait?
	time.Sleep(1 * time.Second)

	// Perform a similarity search
	stage := bson.D{
		// Name of Atlas Vector Search Index tied to Collection

		{"index", testIdxName},
		// Field in Collection containing embedding vectors
		{"path", "plot_embedding"},
		{"queryVector", fooVector},

		// List of embedding vector
		{"numCandidates", 150},
		{"limit", 10},
	}

	found, err := SearchVectors(context.Background(), coll, stage)
	if err != nil {
		log.Fatalf("failed to search vectors: %v", err)
	}

	scores := []float64{}
	for _, fdoc := range found {
		for _, felement := range fdoc {
			if felement.Key == "score" {
				scores = append(scores, felement.Value.(float64))
			}
		}
	}

	fmt.Println(len(found), scores[0], math.Abs(scores[1]-0.0039) < 1e-3)
	// Output: 2 1 true

}
