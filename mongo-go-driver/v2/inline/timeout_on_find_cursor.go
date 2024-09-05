//go:build timeoutOnFindCursor

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
	client, err := mongo.Connect(options.Client().SetMonitor(newCM()))
	if err != nil {
		panic(err)
	}

	defer client.Disconnect(context.Background())

	// Insert a bunch of data to query
	toInsert := []bson.D{}
	for i := 0; i < 100; i++ {
		toInsert = append(toInsert, bson.D{{"x", i}})
	}

	coll := client.Database("db").Collection("coll")
	defer coll.Drop(context.Background())

	if _, err := coll.InsertMany(context.Background(), toInsert); err != nil {
		panic(err)
	}

	// Search with small batch size
	findOpts := options.Find().SetBatchSize(1)
	cur, err := coll.Find(context.Background(), bson.D{}, findOpts)
	if err != nil {
		panic(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	d := []bson.D{}
	if err = cur.All(ctx, &d); err != nil {
		panic(err)
	}
}

func newCM() *event.CommandMonitor {
	return &event.CommandMonitor{
		Started: func(ctx context.Context, cse *event.CommandStartedEvent) {
			if cse.CommandName != "getMore" {
				return
			}

			fmt.Println(cse.Command)
		},
	}
}
