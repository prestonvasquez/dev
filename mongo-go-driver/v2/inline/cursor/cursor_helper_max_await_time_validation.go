package cursor

//
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
//	client, err := mongo.Connect(options.Client())
//	if err != nil {
//		log.Fatal("failed to connect to server: %v", err)
//	}
//
//	defer func() {
//		if err = client.Disconnect(context.Background()); err != nil {
//			log.Fatal("failed to disconnect client: %v", err)
//		}
//	}()
//
//	// Specify the database and collection
//	db := client.Database("exampleDB")
//	coll := db.Collection("exampleCappedCollection")
//
//	opts := options.Find().SetMaxAwaitTime(10 * time.Second)
//
//	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
//	defer cancel()
//
//	_, err = coll.Find(ctx, bson.D{}, opts)
//	fmt.Printf("expect validation error, got: %v\n", err != nil)
//}
