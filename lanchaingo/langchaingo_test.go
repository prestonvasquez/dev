package langchaingo

import (
	"context"
	"testing"

	"github.com/tmc/langchaingo/embeddings"
	"github.com/tmc/langchaingo/schema"
	"go.mongodb.org/mongo-driver/bson"
)

func Test_AddDocuments(t *testing.T) {
	// Example taken from langchain mongodb vectorstore.
	// https://tinyurl.com/5c5mt724
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

	// Extract the body of text for embedding.
	texts := make([]string, 0, len(docs))
	for _, doc := range docs {
		texts = append(texts, doc.PageContent)
	}

	// Create a mock LLM object.
	llm := &mockOpenAILLM{}

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

	// From here you just insert the above data into an atlas cluster.
}
