package main

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"log"
	"math/rand/v2"
	"strings"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/event"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func main() {
	ctx := context.Background()

	client, err := mongo.Connect(options.Client().SetMonitor(monitor()))
	handle(err)

	db := client.Database("test_big")
	coll := db.Collection("testcoll")

	err = db.Drop(ctx)
	handle(err)

	writes := makeWrites()

	res, err := coll.BulkWrite(ctx, writes)
	handle(err)

	fmt.Println(res.UpsertedCount)
}

func handle(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func makeWrites() []mongo.WriteModel {
	var writes []mongo.WriteModel

	for i := range 9 {
		doc := makeBigDoc(i)

		writes = append(
			writes,
			mongo.NewReplaceOneModel().
				SetFilter(bson.D{{"_id", doc.ID}}).
				SetReplacement(doc).
				SetUpsert(true),
		)
	}

	return writes
}

type myDoc struct {
	ID    string
	Value int64
}

func makeBigDoc(id int) myDoc {
	idSize := 2 * 1024 * 1024 // 2 MB
	md5Length := 32           // byte length of hex string representation of md5

	// We need a unique string with a fixed length, where that fixed length is a multiple of our
	// large ID size constants above. MD5 is 16 bytes, which works.
	idHash := md5.Sum([]byte(fmt.Sprintf("%d", id)))

	return myDoc{
		ID:    strings.Repeat(hex.EncodeToString(idHash[:]), idSize/md5Length),
		Value: rand.Int64(),
	}
}

func monitor() *event.CommandMonitor {
	return &event.CommandMonitor{
		Started: func(_ context.Context, evt *event.CommandStartedEvent) {
			log.Printf("command started: %s", evt.CommandName)
		},
	}
}
