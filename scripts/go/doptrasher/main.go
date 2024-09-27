package main

import (
	"context"
	"log"
	"os"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	// Check if the correct number of arguments are provided
	if len(os.Args) < 2 {
		log.Fatalf("Usage: %s <mongodb-uri>", os.Args[0])
	}

	// Get the MongoDB URI from the arguments
	uri := os.Args[1]

	// Create a new client and connect to the server
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(uri))
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	defer func() { _ = client.Disconnect(context.Background()) }()

	// Specify the database and bucket name
	db := client.Database("diskhop")
	bucketName := "trash"

	// Collection names for GridFS
	filesCollection := bucketName + ".files"
	chunksCollection := bucketName + ".chunks"

	// Delete all files and chunks from the bucket
	err = deleteAllFilesAndChunks(db, filesCollection, chunksCollection)
	if err != nil {
		log.Fatalf("Failed to delete files and chunks: %v", err)
	}

	// Compact the files collection
	err = compactCollection(db, filesCollection)
	if err != nil {
		log.Fatalf("Failed to compact files collection: %v", err)
	}

	// Compact the chunks collection
	err = compactCollection(db, chunksCollection)
	if err != nil {
		log.Fatalf("Failed to compact chunks collection: %v", err)
	}

	log.Println("Bucket compacted successfully")
}

// deleteAllFilesAndChunks deletes all files and chunks from the specified collections
func deleteAllFilesAndChunks(db *mongo.Database, filesCollection, chunksCollection string) error {
	// Delete all files from the files collection
	_, err := db.Collection(filesCollection).DeleteMany(context.Background(), bson.D{})
	if err != nil {
		return err
	}

	// Delete all chunks from the chunks collection
	_, err = db.Collection(chunksCollection).DeleteMany(context.Background(), bson.D{})
	if err != nil {
		return err
	}

	log.Printf("All files and chunks deleted from collections %s and %s", filesCollection, chunksCollection)
	return nil
}

// compactCollection runs the compact command on a given collection
func compactCollection(db *mongo.Database, collectionName string) error {
	command := bson.D{
		{"compact", collectionName},
		{"force", true},
	}

	var result bson.M
	err := db.RunCommand(context.Background(), command).Decode(&result)
	if err != nil {
		return err
	}

	log.Printf("Compaction result for collection %s: %v", collectionName, result)
	return nil
}
