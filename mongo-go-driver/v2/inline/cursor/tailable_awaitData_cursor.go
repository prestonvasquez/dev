package cursor

import (
	"context"
	"log"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// package main
// import (
//
//	"context"
//	"fmt"
//	"log"
//	"time"
//
//	"go.mongodb.org/mongo-driver/v2/bson"
//	"go.mongodb.org/mongo-driver/v2/event"
//	"go.mongodb.org/mongo-driver/v2/mongo"
//	"go.mongodb.org/mongo-driver/v2/mongo/options"
//
// )
//
//	func main() {
//		client, err := mongo.Connect(options.Client().SetMonitor(newCommandMonitor()))
//		if err != nil {
//			log.Fatal("failed to connect to server: %v", err)
//		}
//
//		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
//		defer cancel()
//
//		defer func() {
//			if err = client.Disconnect(ctx); err != nil {
//				log.Fatal("failed to disconnect client: %v", err)
//			}
//		}()
//
//		// Specify the database and collection
//		db := client.Database("exampleDB")
//		collection := db.Collection("exampleCappedCollection")
//
//		// Ensure the collection is capped (tailable cursors only work on capped collections)
//		err = setupCappedCollection(ctx, collection)
//		if err != nil {
//			log.Fatal(err)
//		}
//
//		docErrChCtx, docErrChCtxCancel := context.WithCancel(context.Background())
//		docErrCh := startCreatingDocuments(docErrChCtx, collection)
//
//		// Configure the tailable cursor
//		findOptions := options.Find()
//		findOptions.SetCursorType(options.TailableAwait)
//		findOptions.SetBatchSize(1)
//		findOptions.SetMaxAwaitTime(5 * time.Millisecond)
//
//		// Start watching for changes in the capped collection
//		cursor, err := collection.Find(context.Background(), bson.M{}, findOptions)
//		if err != nil {
//			log.Fatal("Error creating tailable cursor:", err)
//		}
//		defer cursor.Close(ctx)
//
//		fmt.Println("Listening for new documents...")
//
//		nextCtx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
//		defer cancel()
//
//		iterations := 0
//		for {
//			iterations++
//			fmt.Println("calling next...")
//			startTime := time.Now()
//			if cursor.Next(nextCtx) {
//				doc := bson.M{}
//				err := cursor.Decode(&doc)
//				if err != nil {
//					// Error decoding document
//					log.Println("Error decoding document:", err)
//					continue
//				}
//				// Process the document
//				fmt.Println("Received document:", doc)
//			} else {
//				// If the cursor is exhausted, check for errors or exit the loop
//				if err := cursor.Err(); err != nil {
//					log.Println("Cursor error:", err)
//					break
//				}
//				time.Sleep(1 * time.Second) // Wait briefly before continuing
//			}
//			elapsed := time.Since(startTime)
//			fmt.Println("elapsed: ", elapsed)
//		}
//
//		fmt.Println("iterations:", iterations)
//
//		docErrChCtxCancel()
//		if err := <-docErrCh; err != nil {
//			log.Fatalf("creating docs failed: %v", err)
//		}
//
//		fmt.Println("Stopped listening for new documents.")
//	}
//
// Helper function to set up a capped collection (if not already existing)
func setupCappedCollection(ctx context.Context, collection *mongo.Collection) error {
	db := collection.Database()
	collectionName := collection.Name()

	// Check if the collection exists
	collections, err := db.ListCollectionNames(ctx, bson.M{"name": collectionName})
	if err != nil {
		return err
	}

	if len(collections) == 0 {
		// Collection does not exist, create it as a capped collection
		cappedOpts := options.CreateCollection().SetCapped(true).SetSizeInBytes(1024 * 1024) // 1 MB size
		err = db.CreateCollection(ctx, collectionName, cappedOpts)
		if err != nil {
			return err
		}
		log.Println("Capped collection created:", collectionName)
	} else {
		log.Println("Capped collection already exists:", collectionName)
	}

	return nil
}

//
//func startCreatingDocuments(ctx context.Context, coll *mongo.Collection) <-chan error {
//	errChan := make(chan error, 1)
//	go func() {
//		close(errChan)
//
//		for {
//			_, err := coll.InsertOne(ctx, bson.D{{Key: "x", Value: 1}})
//			if err != nil {
//				errChan <- err
//
//				break
//			}
//
//			time.Sleep(1 * time.Second)
//		}
//	}()
//
//	return errChan
//}
//
//func newCommandMonitor() *event.CommandMonitor {
//	return &event.CommandMonitor{
//		Started: func(_ context.Context, cse *event.CommandStartedEvent) {
//			if cse.CommandName == "find" {
//				log.Println("find: ", cse.Command)
//			}
//		},
//	}
//}
