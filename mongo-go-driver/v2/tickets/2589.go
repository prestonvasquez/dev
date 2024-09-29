package main

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

// end-restaurant-struct

func main() {
	client, err := mongo.Connect()
	if err != nil {
		panic(err)
	}

	defer func() {
		if err = client.Disconnect(context.TODO()); err != nil {
			panic(err)
		}
	}()

	coll := client.Database("test").Collection("coll")

	coll.Drop(context.Background())

	res := coll.FindOne(context.Background(), bson.D{})
	fmt.Println("Find One:", res.Err()) // Returns no document

	cursor, err := coll.Find(context.TODO(), bson.D{})
	if err != nil {
		panic(err)
	}

	var results []any
	if err = cursor.All(context.TODO(), &results); err != nil {
		panic(err)
	}

	fmt.Println(results == nil, len(results))
}
