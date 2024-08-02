package main

import (
	"context"
	"devv1/util/indexutil"
	"fmt"
	"log"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	opts := options.Client().ApplyURI("mongodb://localhost:27017")

	client, err := mongo.Connect(context.Background(), opts)
	if err != nil {
		log.Fatalf("failed to create client: %v", err)
	}

	defer client.Disconnect(context.Background())

	// Create some indexes
	coll := client.Database("db").Collection("coll")
	defer coll.Drop(context.Background())

	if err := indexutil.CreateN(context.Background(), coll, 4); err != nil {
		log.Fatalf("failed to create indexes: %v", err)
	}

	cur, err := coll.Indexes().List(context.Background())
	if err != nil {
		panic(err)
	}

	count := 0
	for cur.Next(context.Background()) {
		count++
	}

	log.Printf("num of indexes: %v", count)

	res, err := coll.Indexes().DropAll(context.Background())
	if err != nil {
		log.Fatalf("failed to drop indexes: %v", err)
	}

	type dropResult struct {
		NIndexesWas int
	}

	dres := dropResult{}
	if err := bson.Unmarshal(res, &dres); err != nil {
		log.Fatalf("failed to decode: %v", err)
	}

	fmt.Println(dres.NIndexesWas)
}
