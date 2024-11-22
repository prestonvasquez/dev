package main

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/writeconcern"
)

func main() {
	ctx := context.Background()
	opts := options.
		Client().
		ApplyURI("mongodb://localhost:27017").
		SetWriteConcern(writeconcern.Unacknowledged())

	client, err := mongo.Connect(ctx, opts)
	if err != nil {
		panic(err)
	}

	defer func() { _ = client.Disconnect(ctx) }()

	coll := client.Database("x").Collection("y")

	ioOpts := options.InsertOne().SetBypassDocumentValidation(true)

	_, err = coll.InsertOne(ctx, bson.D{{"x", 1}}, ioOpts)
	fmt.Println("err: ", err)
}
