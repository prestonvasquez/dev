package cursor

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/event"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

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

func startCreatingDocuments(ctx context.Context, coll *mongo.Collection) <-chan error {
	errChan := make(chan error, 1)
	go func() {
		close(errChan)

		for {
			_, err := coll.InsertOne(ctx, bson.D{{Key: "x", Value: 1}})
			if err != nil {
				errChan <- err

				break
			}

			time.Sleep(1 * time.Second)
		}
	}()

	return errChan
}

func newCommandMonitor() *event.CommandMonitor {
	return &event.CommandMonitor{
		Started: func(_ context.Context, cse *event.CommandStartedEvent) {
			if cse.CommandName == "find" {
				log.Println("find: ", cse.Command)
			}
			if cse.CommandName == "getMore" {
				log.Println("getMore: ", cse.Command)
			}
		},
	}
}
