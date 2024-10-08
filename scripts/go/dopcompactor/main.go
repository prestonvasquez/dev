package main

import (
	"context"
	"fmt"
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
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(uri).SetDirect(true))
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	defer func() { _ = client.Disconnect(context.Background()) }()

	dbs := []string{"vid", "diskhop"}

	for _, db := range dbs {
		log.Printf("Compacting GridFS buckets in database: %s", db)

		// Specify the database
		db := client.Database(db)

		// List all GridFS buckets and compact them, excluding "trash"
		err = compactAllBuckets(db, "trash")
		if err != nil {
			log.Fatalf("Failed to compact buckets: %v", err)
		}
	}

	log.Println("Buckets compacted successfully")
}

// compactAllBuckets compacts all GridFS buckets in the database, except for the excluded bucket
func compactAllBuckets(db *mongo.Database, excludeBucket string) error {
	// List collections in the database
	collections, err := db.ListCollectionNames(context.Background(), bson.D{})
	if err != nil {
		return err
	}

	for _, collection := range collections {
		log.Printf("Compacting collection: %s", collection)
		if collection == excludeBucket+".files" || collection == excludeBucket+".chunks" {
			continue
		}

		if err := compactCollection(db, collection); err != nil {
			return fmt.Errorf("failed to compact files collection: %v", err)
		}
	}

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
