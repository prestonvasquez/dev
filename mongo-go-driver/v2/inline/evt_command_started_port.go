package main

import (
	"context"
	"fmt"
	"net"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/event"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func main() {
	opts := options.Client().SetMonitor(newCommandMonitor())

	client, err := mongo.Connect(opts)
	if err != nil {
		panic(err)
	}

	defer func() { _ = client.Disconnect(context.Background()) }()

	coll := client.Database("db").Collection("coll")
	coll.Drop(context.Background())

	_, err = coll.InsertOne(context.Background(), bson.D{{"x", 1}})
	if err != nil {
		panic(err)
	}
}

func newCommandMonitor() *event.CommandMonitor {
	return &event.CommandMonitor{
		Started: func(ctx context.Context, cse *event.CommandStartedEvent) {

			host, port, err := net.SplitHostPort(cse.ConnectionID)
			fmt.Println("connectionID", host, port, err)
		},
	}
}
