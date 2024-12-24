package main

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func main() {
	opts := options.Client().ApplyURI("mongodb://localhost:27017")

	// If the username is "x", use password "z"
	hsetter := func(opts *options.ClientOptions) error {
		td := 100 * time.Millisecond
		opts.HeartbeatInterval = &td

		return nil
	}

	opts.Opts = append(opts.Opts, hsetter)

	client, err := mongo.Connect(opts)
	if err != nil {
		panic(err)
	}

	defer func() { _ = client.Disconnect(context.Background()) }()
}
