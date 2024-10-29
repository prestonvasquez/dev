package main

import (
	"context"
	"fmt"
	"log"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/event"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/x/mongo/driver"
	"go.mongodb.org/mongo-driver/v2/x/mongo/driver/topology"
)

func createBlockFP(client *mongo.Client, cmd string, blockTime, iter int) (func(), error) {
	admindb := client.Database("admin")

	// Create a document for the run command that sets a fail command that is always on.
	failCommand := bson.D{
		{Key: "configureFailPoint", Value: "failCommand"},
		{Key: "mode", Value: bson.D{{"times", iter}}},
		{Key: "data", Value: bson.D{
			{Key: "blockConnection", Value: true},
			{Key: "blockTimeMS", Value: blockTime},
			{Key: "failCommands", Value: bson.A{cmd}},
		}},
	}

	err := admindb.RunCommand(context.Background(), failCommand).Err()
	if err != nil {
		return func() {}, err
	}

	return func() {
		doc := bson.D{
			{Key: "configureFailPoint", Value: "failCommand"},
			{Key: "mode", Value: "off"},
		}

		err = admindb.RunCommand(context.Background(), doc).Err()
		if err != nil {
			log.Fatalf("could not disable fail point command: %v", err)
		}
	}, nil
}

func main() {
	commandMonitor := &event.CommandMonitor{
		Started: func(ctx context.Context, cse *event.CommandStartedEvent) {
			if cse.CommandName == "insert" {
				fmt.Println("started: ", cse.Command)
			}
		},
		Failed: func(ctx context.Context, cse *event.CommandFailedEvent) {
			if cse.CommandName == "insert" {
				derr, _ := cse.Failure.(driver.Error)
				cerr, _ := derr.Wrapped.(topology.ConnectionError)
				fmt.Printf("wrapped type: %T\n", cerr.Wrapped)
			}
		},
	}

	// Set up MongoDB client options
	opts := options.Client().SetMaxPoolSize(1).SetMonitor(commandMonitor)

	client, err := mongo.Connect(opts)
	if err != nil {
		log.Fatalf("failed to connect to server: %v", err)
	}
	defer func() {
		_ = client.Disconnect(context.Background())
	}()

	// Create a failpoint that will block 1 time for 750ms.
	teardown, err := createBlockFP(client, "insert", 750, 1)
	if err != nil {
		log.Fatalf("failed to create blocking failpoint: %v", err)
	}
	defer teardown()

	db := client.Database("db")

	cmd := bson.D{
		{"insert", "coll"},
		{"documents", bson.A{bson.D{{"y", 1}}}},
		{"maxTimeMS", 50},
	}

	res := db.RunCommand(context.Background(), cmd)
	if err := res.Err(); err != nil {
		fmt.Println("err: ", err)
	}

	//res = db.RunCommand(context.Background(), cmd)
	//if err := res.Err(); err != nil {
	//	fmt.Println("err: ", err)
	//}

	// Attempt to insert with a short timeout
	//_, err = coll.InsertOne(ctx, bson.D{})
	//cerr, ok := err.(mongo.CommandError)
	//if ok {
	//	fmt.Println("raw: ", len(cerr.Raw))
	//}

	//if errors.Is(err, context.DeadlineExceeded) {
	//	log.Printf("Insert failed as expected: %v", err)
	//}

	//// Attempt to insert with a longer timeout
	//ctx, cancel = context.WithTimeout(context.Background(), 500*time.Millisecond)
	//defer cancel()

	//_, err = coll.InsertOne(ctx, bson.D{})
	//if errors.Is(err, context.DeadlineExceeded) {
	//	log.Printf("Insert failed as expected: %v", err)
	//} else {
	//	log.Println("Insert succeeded unexpectedly")
	//}
}
