package main

import (
	"context"
	"crypto/tls"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	// Set up the MongoDB client with TLS options
	tlsConfig := &tls.Config{
		InsecureSkipVerify: true, // For testing purposes only; do not use in production
	}
	clientOptions := options.Client().
		SetTLSConfig(tlsConfig)

	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer client.Disconnect(context.TODO())

	// Start the connection prober in a separate goroutine
	go func() {
		collection := client.Database("testdb").Collection("testcollection")
		for {
			_, err := collection.InsertOne(context.TODO(), map[string]interface{}{"test": "data"})
			if err != nil {
				log.Printf("Insert failed: %v", err)
			} else {
				log.Println("Insert succeeded")
			}
			time.Sleep(10 * time.Second)
		}
	}()

	// Simulate changing the server's TLS settings after some time
	time.Sleep(30 * time.Second) // Wait for a few inserts
	log.Println("Changing server TLS settings from preferSSL to requireSSL...")

	// Here you would implement the logic to change the server's TLS settings.
	// This is a placeholder for the actual server configuration change.
	// In a real scenario, you would need to restart the MongoDB server or change its configuration.

	// Wait for a while to observe the behavior
	time.Sleep(60 * time.Second)
	log.Println("Finished testing.")
}

//func main() {
//	opts := options.Client().SetMonitor(monitor())
//
//	client, err := mongo.Connect(opts)
//	if err != nil {
//		panic(err)
//	}
//
//	defer func() { _ = client.Disconnect(context.Background()) }()
//
//	teardown, err := configureBlockingFP(context.Background(), client)
//	if err != nil {
//		panic(err)
//	}
//
//	defer teardown()
//
//	coll := client.Database("db").Collection("coll")
//
//	_, err = coll.InsertOne(context.Background(), bson.D{})
//	if err != nil {
//		panic(err)
//	}
//}
//
//func configureBlockingFP(ctx context.Context, client *mongo.Client) (func(), error) {
//	admindb := client.Database("admin")
//
//	// Create a document for the run command that sets a fail command that is always on.
//	failCommand := bson.D{
//		{"configureFailPoint", "failCommand"},
//		{"mode", bson.D{{"times", 1}}},
//		{"data",
//			bson.D{
//				{"failCommands", bson.A{"insert"}},
//				{"errorCode", 11602},
//				{"closeConnection", false},
//			},
//		},
//	}
//
//	err := admindb.RunCommand(ctx, failCommand).Err()
//	if err != nil {
//		return func() {}, err
//	}
//
//	return func() {
//		doc := bson.D{
//			{"configureFailPoint", "failCommand"},
//			{"mode", "off"},
//		}
//		err = admindb.RunCommand(ctx, doc).Err()
//		if err != nil {
//			log.Fatalf("could not disable fail point command: %v", err)
//		}
//	}, nil
//}
//
//func monitor() *event.CommandMonitor {
//	return &event.CommandMonitor{
//		Started: func(_ context.Context, cse *event.CommandStartedEvent) {
//			if cse.CommandName == "insert" {
//				fmt.Println("started: ", cse)
//			}
//		},
//		Succeeded: func(_ context.Context, cse *event.CommandSucceededEvent) {
//			if cse.CommandName == "insert" {
//				fmt.Println("succeeded: ", cse)
//			}
//		},
//		Failed: func(_ context.Context, cfe *event.CommandFailedEvent) {
//			if cfe.CommandName == "insert" {
//				fmt.Println("failed: ", cfe)
//			}
//		},
//	}
//}
