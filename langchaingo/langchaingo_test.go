package langchaingo

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tmc/langchaingo/embeddings"
	"github.com/tmc/langchaingo/schema"
	"go.mongodb.org/mongo-driver/bson"
)

// Example taken from langchain mongodb vectorstore.
// https://tinyurl.com/5c5mt724
func newExampleDocs() []schema.Document {
	docs := []schema.Document{
		{
			PageContent: "foo",
			Metadata:    map[string]any{"baz": "bar"},
		},
		{
			PageContent: "thud",
			Metadata:    map[string]any{"baz": "bar"},
		},
		{
			PageContent: "i will be deleted :(",
		},
	}

	return docs
}

func Test_AddDocuments(t *testing.T) {
	docs := newExampleDocs()

	// Extract the body of text for embedding.
	texts := make([]string, 0, len(docs))
	for _, doc := range docs {
		texts = append(texts, doc.PageContent)
	}

	// Create a mock LLM object.
	llm := &MockOpenAILLM{}

	// Use the OpenAI LLM object to create an embedder.
	openAIEmbedder, err := embeddings.NewEmbedder(llm)
	if err != nil {
		t.Fatalf("failed to construct an embedder for an openAI LLM: %v", err)
	}

	// Create vectors from the embedding.
	vectors, err := openAIEmbedder.EmbedDocuments(context.Background(), texts)
	if err != nil {
		t.Fatalf("failed to embed texts using openAI embedder: %v", err)
	}

	// The number of vectors must be the same as the number of docs.
	if len(vectors) != len(docs) {
		t.Fatalf("number of vectors does not equal the number of docs")
	}

	// Using the vectors, create a slice of bson documents to insert into the
	// Atlas cluster.
	bdocs := make([]bson.D, len(docs))
	for i, doc := range docs {
		bdocs = append(bdocs, bson.D{
			{"text", doc.PageContent},
			{"embedding", vectors[i]},
			{"metadata", doc.Metadata},
		})
	}
}

func TestEmbedder_EmbedQuery(t *testing.T) {
	const query = "thud"

	// Create a mock LLM object.
	llm := &MockOpenAILLM{}

	// Use the OpenAI LLM object to create an embedder.
	openAIEmbedder, err := embeddings.NewEmbedder(llm)
	if err != nil {
		t.Fatalf("failed to construct an embedder for an openAI LLM: %v", err)
	}

	_, err = openAIEmbedder.EmbedQuery(context.Background(), query)
	assert.NoError(t, err, "failed to embed query")
}
