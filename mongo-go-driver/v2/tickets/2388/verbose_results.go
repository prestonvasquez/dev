package main

import (
	"context"
	"fmt"
	"log"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/event"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func main() {
	client, err := mongo.Connect(options.Client().SetMonitor(newMonitor()))
	if err != nil {
		panic(err)
	}

	defer func() { _ = client.Disconnect(context.Background()) }()

	pairs := &mongo.ClientWriteModels{}

	pairs = pairs.AppendInsertOne("db", "coll", mongo.NewClientInsertOneModel().SetDocument(bson.D{{"x", 1}}))
	pairs = pairs.AppendInsertOne("db", "coll2", mongo.NewClientInsertOneModel().SetDocument(bson.D{{"x", 2}}))

	opts := options.ClientBulkWrite().SetVerboseResults(true)

	results, err := client.BulkWrite(context.Background(), pairs, opts)
	if err != nil {
		panic(err)
	}

	fmt.Println(results.InsertResults)
}

func newMonitor() *event.CommandMonitor {
	return &event.CommandMonitor{
		Started: func(_ context.Context, evt *event.CommandStartedEvent) {
			log.Printf("started: %+v\n", evt)
		},
	}
}
