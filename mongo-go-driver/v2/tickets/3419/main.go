package main

import (
	"context"
	"log"
	"os"
	"runtime/pprof"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func main() {
	// Start CPU profile
	f, err := os.Create("cpu.prof")
	if err != nil {
		log.Fatalf("failed to create CPU profile: %v", err)
	}

	pprof.StartCPUProfile(f)
	defer pprof.StopCPUProfile()

	// Connect with maxConnecting=500
	clientOpts := options.Client().SetMaxConnecting(500)

	client, err := mongo.Connect(clientOpts)
	if err != nil {
		log.Fatalf("failed to connect to MongoDB: %v", err)
	}

	defer func() {
		if err := client.Disconnect(context.Background()); err != nil {
			log.Fatalf("failed to disconnect from MongoDB: %v", err)
		}
	}()

	for i := 0; i < 500; i++ {
		if err := client.Ping(context.Background(), nil); err != nil {
			log.Printf("ping failed: %v\n", err)
		}
	}
}
