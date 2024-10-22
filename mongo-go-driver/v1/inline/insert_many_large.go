package main

import (
	"context"
	"fmt"
	"log"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func main() {
	client, err := mongo.Connect(context.Background())
	if err != nil {
		panic(err)
	}

	defer func() { client.Disconnect(context.Background()) }()

	coll := client.Database("db").Collection("coll")

	//docs := make([]bson.D, 1_000)
	//for i := range docs {
	//	docs[i] = bson.D{{"x", i}}
	//}

	docs := []any{}

	docs = append(docs, bson.D{{"y", string(make([]byte, 17_000_000))}})
	docs = append(docs, bson.D{{"y", string(make([]byte, 17_000_000))}})

	_, err = coll.InsertMany(context.Background(), docs)
	if err != nil {
		log.Println("err: ", err)
	}

	cur, err := coll.Find(context.Background(), bson.D{})
	if err != nil {
		log.Fatalf("failed to execute find command: %v", err)
	}

	var found []bson.D

	if err := cur.All(context.Background(), &found); err != nil {
		log.Fatalf("failed to decode data: %v", err)
	}

	fmt.Println("found: ", found)
}
