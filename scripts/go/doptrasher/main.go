package main

import (
	"context"
	"log"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/gridfs"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	// MongoDB connection URI
	// Check if the correct number of arguments are provided
	if len(os.Args) < 2 {
		log.Fatalf("Usage: %s <mongodb-uri>", os.Args[0])
	}

	// Get the MongoDB URI from the first argument
	uri := os.Args[1]

	// Create a new client and connect to the server
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(uri))
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = client.Connect(ctx)
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}

	defer client.Disconnect(ctx)

	// Specify the database and bucket name
	db := client.Database("diskhop")
	bucketName := "trash"

	// Drop the GridFS bucket
	bucket, err := gridfs.NewBucket(db, options.GridFSBucket().SetName(bucketName))
	if err != nil {
		log.Fatalf("Failed to create GridFS bucket: %v", err)
	}

	err = bucket.Drop()
	if err != nil {
		log.Fatalf("Failed to drop GridFS bucket: %v", err)
	}
}
