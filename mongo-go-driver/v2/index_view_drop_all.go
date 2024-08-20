package v2

//
//import (
//	"context"
//	"devv2/util/indexutil"
//	"log"
//
//	"go.mongodb.org/mongo-driver/v2/mongo"
//	"go.mongodb.org/mongo-driver/v2/mongo/options"
//)
//
//func main() {
//	opts := options.Client().ApplyURI("mongodb://localhost:27017")
//
//	client, err := mongo.Connect(opts)
//	if err != nil {
//		log.Fatalf("failed to create client: %v", err)
//	}
//
//	defer client.Disconnect(context.Background())
//
//	// Create some indexes
//	coll := client.Database("db").Collection("coll")
//	defer coll.Drop(context.Background())
//
//	if err := indexutil.CreateN(context.Background(), coll, 4); err != nil {
//		log.Fatalf("failed to create indexes: %v", err)
//	}
//
//	// List the indexes
//	cur, err := coll.Indexes().List(context.Background())
//	if err != nil {
//		panic(err)
//	}
//
//	count := 0
//	for cur.Next(context.Background()) {
//		count++
//	}
//
//	log.Printf("num of indexes: %v", count)
//}
