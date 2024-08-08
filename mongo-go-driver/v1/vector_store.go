package main

import (
	"context"
	"devv1/util/indexutil"
	"log"
	"os"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Store vector embeddings in Atlas cluster

func main() {
	uri := os.Getenv("MONGODB_URI")
	if uri == "" {
		log.Fatalf("MONGOD_URI must be set as an Atlas cluster")
	}

	opts := options.Client().ApplyURI(uri)

	client, err := mongo.Connect(context.Background(), opts)
	if err != nil {
		log.Fatalf("failed to connect: %v", err)
	}

	defer client.Disconnect(context.Background())

	const dbName = "db"
	const collName = "vstore"

	// Create the vectorstore collection
	err = client.Database(dbName).CreateCollection(context.Background(), collName)
	if err != nil {
		log.Fatalf("failed to create vstore collection")
	}

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
	if err != nil {
		log.Fatalf("failed to create vector search index: %v", err)
	}

	// Add some embeddings
}
