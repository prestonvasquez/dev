package cursor

//import (
//	"context"
//	"fmt"
//	"log"
//	"time"
//
//	"go.mongodb.org/mongo-driver/v2/bson"
//	"go.mongodb.org/mongo-driver/v2/mongo"
//	"go.mongodb.org/mongo-driver/v2/mongo/options"
//)
//
//func main() {
//	client, err := mongo.Connect(options.Client().SetHeartbeatInterval(500 * time.Millisecond))
//	if err != nil {
//		log.Fatal("failed to connect to server: %v", err)
//	}
//
//	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
//	defer cancel()
//
//	defer func() {
//		if err = client.Disconnect(ctx); err != nil {
//			log.Fatal("failed to disconnect client: %v", err)
//		}
//	}()
//
//	time.Sleep(2 * time.Second)
//
//	// Specify the database and collection
//	db := client.Database("exampleDB")
//	collection := db.Collection("exampleCappedCollection")
//	collection.Drop(context.Background())
//
//	if err := setupCappedCollection(context.Background(), collection); err != nil {
//		panic(err)
//	}
//
//	collection.InsertOne(context.Background(), bson.D{{"x", 1}})
//
//	// Configure the tailable cursor
//	findOptions := options.Find()
//	findOptions.SetCursorType(options.TailableAwait)
//	findOptions.SetBatchSize(1)
//	findOptions.SetMaxAwaitTime(500 * time.Millisecond)
//
//	findCtx, findCancel := context.WithTimeout(context.Background(), 1*time.Second)
//	defer findCancel()
//
//	// Start watching for changes in the capped collection
//	cursor, err := collection.Find(findCtx, bson.M{}, findOptions)
//	if err != nil {
//		log.Fatal("Error creating tailable cursor:", err)
//	}
//	defer cursor.Close(ctx)
//
//	//ctx, cancel = context.WithTimeout(context.Background(), 500*time.Millisecond)
//	//defer cancel()
//
//	//ctx = context.WithValue(ctx, "latency_context", true)
//	for cursor.Next(context.Background()) {
//		fmt.Println(cursor.Current)
//	}
//
//	//st := time.Now()
//	//var docs []bson.Raw
//	//if err := cursor.All(context.Background(), &docs); err != nil {
//	//	log.Fatalf("cursor failed: %v", err)
//	//}
//
//	//fmt.Println("elapsed: ", time.Since(st))
//}
//
//// Helper function to set up a capped collection (if not already existing)
//func setupCappedCollection(ctx context.Context, collection *mongo.Collection) error {
//	db := collection.Database()
//	collectionName := collection.Name()
//
//	// Check if the collection exists
//	collections, err := db.ListCollectionNames(ctx, bson.M{"name": collectionName})
//	if err != nil {
//		return err
//	}
//
//	if len(collections) == 0 {
//		// Collection does not exist, create it as a capped collection
//		cappedOpts := options.CreateCollection().SetCapped(true).SetSizeInBytes(1024 * 1024) // 1 MB size
//		err = db.CreateCollection(ctx, collectionName, cappedOpts)
//		if err != nil {
//			return err
//		}
//		log.Println("Capped collection created:", collectionName)
//	} else {
//		log.Println("Capped collection already exists:", collectionName)
//	}
//
//	return nil
//}
