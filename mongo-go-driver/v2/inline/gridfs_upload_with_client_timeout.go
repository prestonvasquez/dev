//go:build r1

package main

import (
	"bytes"
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func main() {
	opts := options.Client().SetTimeout(5 * time.Minute)

	client, err := mongo.Connect(opts)
	if err != nil {
		panic(err)
	}

	defer func() { _ = client.Disconnect(context.Background()) }()

	db := client.Database("test")

	bucket := db.GridFSBucket()

	var data bytes.Buffer
	for i := 0; i < 1e7; i++ {
		data.WriteString("some large stuff")
	}

	_, err = bucket.UploadFromStream(context.Background(), "huge_data", bytes.NewBuffer(data.Bytes()))
	if err != nil {
		log.Fatalf("failed to open upload stream: %v", err)
	}
}
