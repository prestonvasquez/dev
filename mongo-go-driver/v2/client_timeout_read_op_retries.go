package main

import (
	"context"
	"devv2/util/eventutil"
	"devv2/util/failpoint"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// Illustrate timeout behavior on operation-level timeouts (context timeouts)
// behavior when a read operation fails.

func main() {
	const insertCmd = "find"
	const errCode = 9001 // SocketException

	// Create a client that logs insert commands
	opts := options.Client().ApplyURI("mongodb://localhost:27017").
		SetMonitor(eventutil.CommandMonitorByName(log.Default(), insertCmd)).
		SetTimeout(5 * time.Second)

	client, err := mongo.Connect(opts)
	if err != nil {
		log.Fatalf("failed to connect: %v", err)
	}

	defer client.Disconnect(context.Background())

	// Create a failpoint that will fail 100x on find.
	closeFP, err := failpoint.NewErrorN(context.Background(), client, insertCmd, errCode, 100)
	if err != nil {
		log.Fatalf("failed to create fail point: %v", err)
	}

	defer closeFP()

	// Run the insert to observe the behavior of a constantly failing read
	// operation with a context timeout.
	coll := client.Database("db").Collection("coll")

	if res := coll.FindOne(context.Background(), bson.D{{"x", 1}}); res.Err() != nil {
		log.Fatalf("failed to insert data: %v", res.Err())
	}
}
