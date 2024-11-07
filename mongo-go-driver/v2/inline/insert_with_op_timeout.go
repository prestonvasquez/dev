package main

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/event"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func main() {
	client, err := mongo.Connect(options.Client().SetMonitor(monitor()))
	if err != nil {
		panic(err)
	}

	defer func() { _ = client.Disconnect(context.Background()) }()

	coll := client.Database("db").Collection("coll")

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	ctx = context.WithValue(ctx, "itest", true)

	_, err = coll.InsertOne(ctx, bson.D{})
	if err != nil {
		panic(err)
	}
}

func monitor() *event.CommandMonitor {
	return &event.CommandMonitor{
		Started: func(ctx context.Context, cse *event.CommandStartedEvent) {
			if cse.CommandName == "insert" {
				fmt.Printf("started: %+v\n", cse)
			}
		},
		Succeeded: func(ctx context.Context, cse *event.CommandSucceededEvent) {
			if cse.CommandName == "insert" {
				fmt.Printf("succeeded: %+v\n", cse)
			}
		},
	}
}
