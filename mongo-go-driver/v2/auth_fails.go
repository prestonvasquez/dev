package v2

//
//import (
//	"context"
//	"log"
//
//	"go.mongodb.org/mongo-driver/v2/mongo"
//	"go.mongodb.org/mongo-driver/v2/mongo/options"
//)
//
//func main() {
//	opts := options.Client().ApplyURI("mongodb://x:y@localhost:27017")
//
//	client, err := mongo.Connect(opts)
//	if err != nil {
//		log.Fatalf("failed to connect: %v", err)
//	}
//
//	defer client.Disconnect(context.Background())
//
//	if err := client.Ping(context.Background(), nil); err != nil {
//		log.Fatalf("failed to ping server: %v", err)
//	}
//}
