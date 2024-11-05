package profiling

import (
	"context"
	"testing"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

// ping function to ping the MongoDB database
func ping(db *mongo.Database, iterations int) {
	command := bson.D{{"ping", 1}}
	for i := 0; i < iterations; i++ {
		db.RunCommand(context.TODO(), command)
	}
}

// BenchmarkPing benchmarks the ping function
func BenchmarkPing(b *testing.B) {
	client, err := mongo.Connect()
	if err != nil {
		b.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer func() {
		if err := client.Disconnect(context.TODO()); err != nil {
			b.Fatalf("Failed to disconnect from MongoDB: %v", err)
		}
	}()

	db := client.Database("sample_restaurants")

	// Run the ping function in parallel
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			ping(db, 100000) // Adjust the number of iterations as needed
		}
	})
}

// To run the benchmark, use the following command in the terminal:
// go test -bench=BenchmarkPing -benchmem
