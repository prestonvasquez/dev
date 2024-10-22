package main

import (
	"context"
	"fmt"
	"log"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/x/mongo/driver/drivertest"
)

func main() {
	md := drivertest.NewMockDeployment()

	opts := options.Client()
	opts.Opts = append(opts.Opts, func(co *options.ClientOptions) error {
		co.Deployment = md

		return nil
	})

	client, err := mongo.Connect(opts)
	if err != nil {
		log.Fatalf("failed to construct client: %v", err)
	}

	mockResponses := []bson.D{
		{
			{Key: "ok", Value: 1},
			{Key: "cursor", Value: bson.D{{Key: "firstBatch", Value: bson.A{bson.D{{"x", 1}}}}}},
		},
	}

	md.AddResponses(mockResponses...)

	coll := client.Database("db").Collection("coll")

	res := coll.FindOne(context.Background(), bson.D{{"x", 1}})
	if res.Err() != nil {
		log.Fatalf("findOne operation failed: %v", res.Err())
	}

	raw, err := res.Raw()
	if err != nil {
		log.Fatalf("failed to get raw value: %v", err)
	}

	fmt.Println("res: ", raw)
}
