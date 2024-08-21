package v2

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"errors"
	"fmt"
	"log"
	"math"
	"os"
	"time"

	"github.com/tmc/langchaingo/embeddings"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// mockLLM will create consistent text embeddings mocking the OpenAI
// text-embedding-3-small algorithm.
type mockLLM struct {
	seen       map[string][]float32
	dimensions int
}

var _ embeddings.EmbedderClient = &mockLLM{}

// n for mocking an embedding.
// createMockEmbedding will create a float32 vector with random elements of size
func createMockEmbedding(n int) []float32 {
	f32s := make([]float32, n)

	for i := range f32s {
		var b [4]byte

		rand.Read(b[:])
		f32s[i] = float32(binary.LittleEndian.Uint32(b[:])) / math.MaxUint32
	}

	return f32s
}

// createEmbedding will return vector embeddings for the mock LLM, maintaining
// consitency.
func (emb *mockLLM) CreateEmbedding(ctx context.Context, texts []string) ([][]float32, error) {
	if emb.seen == nil {
		emb.seen = map[string][]float32{}
	}

	vectors := make([][]float32, len(texts))
	for i, text := range texts {
		if f32s := emb.seen[text]; len(f32s) > 0 {
			vectors[i] = f32s

			continue
		}

		f32s := createMockEmbedding(emb.dimensions)

		vectors[i] = f32s
		emb.seen[text] = f32s // ensure consistency
	}

	return vectors, nil
}

// vectorField defines the fields of an index used for vector search.
type vectorField struct {
	Type          string `bson:"type"`
	Path          string `bson:"path"`
	NumDimensions int    `bson:"numDimensions"`
	Similarity    string `bson:"similarity"`
}

// createVectorSearchIndex will create a vector search index on the "db.vstore"
// collection named "vector_index" with the provided field. This function blocks
// until the index has been created.
func createVectorSearchIndex(
	ctx context.Context,
	coll *mongo.Collection,
	idxName string,
	field vectorField,
) (string, error) {
	def := struct {
		Fields []vectorField `bson:"fields"`
	}{
		Fields: []vectorField{field},
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

	return searchName, nil
}

// dropVectorSearchIndex will attempt to drop the search index by name, awaiting
// that it has been dropped. This function blocks until the index has been
// dropped.
func dropVectorSearchIndex(ctx context.Context, coll *mongo.Collection, idxName string) error {
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

func runVectorSearchExample(ctx context.Context, query string, toInsert []string) ([]bson.D, error) {
	const (
		dbName   = "db"
		collName = "vstore"
		index    = "vector_index"
	)

	var coll *mongo.Collection

	log.Println("creating collection...")
	{
		uri := os.Getenv("MONGODB_URI")
		if uri == "" {
			return nil, errors.New("MONGOD_URI must be set as an Atlas cluster")
		}

		client, err := mongo.Connect(options.Client().ApplyURI(uri))
		if err != nil {
			return nil, fmt.Errorf("failed to connect: %w", err)
		}

		err = client.Database(dbName).CreateCollection(ctx, collName)
		if err != nil {
			return nil, fmt.Errorf("failed to create collection: %w", err)
		}

		coll = client.Database(dbName).Collection(collName)
		defer coll.Drop(ctx)
	}

	log.Println("dropping and creating the index...")
	{
		_ = dropVectorSearchIndex(ctx, coll, index)

		field := vectorField{
			Type:          "vector",
			Path:          "plot_embedding",
			NumDimensions: 1536,
			Similarity:    "euclidean",
		}

		_, err := createVectorSearchIndex(ctx, coll, index, field)
		if err != nil {
			log.Fatalf("failed to create vector search index: %v", err)
		}
	}

	embedder, err := embeddings.NewEmbedder(&mockLLM{dimensions: 1536})
	if err != nil {
		return nil, fmt.Errorf("failed to construct an embedder for an openAI LLM: %w", err)
	}

	log.Println("embedding and inserting documents...")
	{
		// Create vectors from the embedding.
		vectors, err := embedder.EmbedDocuments(ctx, toInsert)
		if err != nil {
			log.Fatalf("failed to embed texts using openAI embedder: %v", err)
		}

		docs := []interface{}{}
		for i, ti := range toInsert {
			docs = append(docs, bson.D{
				{Key: "page_content", Value: ti},
				{Key: "plot_embedding", Value: vectors[i]},
			})
		}

		_, err = coll.InsertMany(ctx, docs)
		if err != nil {
			log.Fatalf("failed to insert documents: %v", err)
		}
	}

	// TODO: Why do we have to wait?
	time.Sleep(1 * time.Second)

	found := []bson.D{}

	log.Println("executing similarity search...")
	{
		// 5. Perform a similarity search. This will perform a similarity search on
		// the vector space. A basic similarity search should look something like
		// this:
		//
		// vector_store.similarity_search(query="thud",k=1)
		//
		// The user will provide a query. We need to get the embedding and then add
		// that to the aggregation pipeline's stage.
		//
		// In practice, we would use the embeddings.Embedder to do this:
		//
		// openAIEmbedder, _ := embeddings.NewEmbedder(llm)
		// vector, _ := openAIEmbedder.EmbedQuery(ctx, query)
		queryVector, err := embedder.EmbedQuery(ctx, query)
		if err != nil {
			return nil, fmt.Errorf("failed to embed query: %w", err)
		}

		stage := bson.D{
			// Name of Atlas Vector Search Index tied to Collection

			{"index", index},
			// Field in Collection containing embedding vectors
			{"path", "plot_embedding"},
			{"queryVector", queryVector},

			// List of embedding vector
			{"numCandidates", 150},
			{"limit", 10},
		}

		// Create a vector search pipeline
		pipeline := mongo.Pipeline{
			bson.D{
				{"$vectorSearch", stage},
			},
			bson.D{
				{"$set", bson.D{{"score", bson.D{{"$meta", "vectorSearchScore"}}}}},
			},
		}

		cur, err := coll.Aggregate(ctx, pipeline)
		if err != nil {
			return nil, fmt.Errorf("failed to aggregate: %w", err)
		}

		for cur.Next(ctx) {
			var d bson.D
			err := cur.Decode(&d)
			if err != nil {
				return nil, fmt.Errorf("failed to decode: %w", err)
			}

			found = append(found, d)
		}
	}

	return found, nil
}
