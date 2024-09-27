package main

import (
	"context"
	"fmt"
	"reflect"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

func main() {
	client, err := mongo.Connect()
	if err != nil {
		panic(err)
	}

	defer func() { _ = client.Disconnect(context.Background()) }()

	err = client.Database("t3341").CreateCollection(context.Background(), "coll")
	if err != nil {
		panic(err)
	}

	coll := client.Database("t3341").Collection("coll")
	_ = coll.Drop(context.Background())

	doc := map[string]any{
		"_id":    "zen",
		"logger": map[string]any{"level": "trace", "output": "journal"},
	}

	if _, err := coll.InsertOne(context.Background(), doc); err != nil {
		panic(err)
	}

	res := coll.FindOne(context.Background(), bson.M{"_id": "zen"})
	if err := res.Err(); err != nil {
		panic(err)
	}

	var data map[string]any
	if err := res.Decode(&data); err != nil {
		panic(err)
	}

	fmt.Println(reflect.TypeOf(data["logger"]))
}
