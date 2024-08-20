package v2

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"log"
	"math"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// createMockEmbedding will create a float32 vector with random elements of size
// n for mocking an embedding.
func createMockEmbedding(n int) []float32 {
	f32s := make([]float32, n)

	for i := range f32s {
		var b [4]byte

		rand.Read(b[:])
		f32s[i] = float32(binary.LittleEndian.Uint32(b[:])) / math.MaxUint32
	}

	return f32s
}

// Convert ClientOptionsBuilder to ClientOptions
func ExampleClientOptionBuilderToClientOptions() {
	var opts options.ClientOptions
	for _, set := range options.Client().ApplyURI("mongodb://localhost:27017/?appName=foo").Opts {
		if err := set(&opts); err != nil {
			log.Fatalf("failed to set: %v", err)
		}
	}

	fmt.Println(*opts.AppName)
	// Output: foo
}

func ExampleInsertMany_F32ArrayElement() {
	client, err := mongo.Connect()
	if err != nil {
		log.Fatalf("failed to connect: %v", err)
	}

	coll := client.Database("db").Collection("coll")

	bdocs := make([]bson.D, 10)
	for i := 0; i < 10; i++ {
		bdocs[i] = bson.D{{"vec", createMockEmbedding(1536)}}
	}

	_, err = coll.InsertMany(context.TODO(), bdocs)
	fmt.Println(err)
	// Output: <nil>
}

func ExampleInsertMany_EmptyDocs() {
	client, err := mongo.Connect()
	if err != nil {
		log.Fatalf("failed to connect: %v", err)
	}

	coll := client.Database("db").Collection("coll")
	_, err = coll.InsertMany(context.TODO(), []bson.D{{}})

	fmt.Println(err)
	// Output: <nil>
}

func ExampleInsertMany_DocWithEmptyValue() {
	client, err := mongo.Connect()
	if err != nil {
		log.Fatalf("failed to connect: %v", err)
	}

	bdocs := append([]bson.D{}, bson.D{{Key: "text"}})

	coll := client.Database("db").Collection("coll")
	_, err = coll.InsertMany(context.TODO(), bdocs)

	fmt.Println(err)
	// Output: <nil>
}

// Set an operation-level timeout and see if it sets a server-side timeout.
func ExampleOpLevelTimeout() {
	client, err := mongo.Connect(options.Client().SetMonitor(commandMonitorByName(log.Default(), "insert")))
	if err != nil {
		log.Fatalf("failed to connect: %v", err)
	}

	defer client.Disconnect(context.Background())

	coll := client.Database("db").Collection("coll")
	defer coll.Drop(context.Background())

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if _, err := coll.InsertOne(ctx, bson.D{{"x", 1}}); err != nil {
		log.Fatalf("failed to insert: %v", err)
	}

	fmt.Println("done")
	// Output: done
}
