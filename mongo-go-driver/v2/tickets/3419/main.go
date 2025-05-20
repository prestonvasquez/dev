package main

import (
	"context"
	"log"
	"os"
	"runtime/pprof"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

const (
	poolSize    = 100
	workerCount = 100
)

func main() {
	uri := os.Getenv("MONGODB_URI")
	if uri == "" {
		uri = "mongodb://localhost:27017"
	}

	// 1) Connect with a sane pool
	client, err := mongo.Connect(options.Client().
		ApplyURI(uri).
		SetMaxPoolSize(poolSize).
		SetMaxConnecting(poolSize).
		SetServerSelectionTimeout(5 * time.Second).
		SetConnectTimeout(5 * time.Second))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(context.Background())

	// 2) Warm up the pool
	for i := 0; i < poolSize; i++ {
		if err := client.Ping(context.Background(), nil); err != nil {
			log.Fatalf("warm-up ping failed: %v", err)
		}
	}
	log.Printf("Pool warmed to %d connections\n", poolSize)

	// 3) Spawn workers that reuse those connections
	for i := 0; i < workerCount; i++ {
		go func(id int) {
			for {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				err := client.Ping(ctx, nil)
				cancel()
				if err != nil {
					log.Printf("worker %d ping error: %v", id, err)
					time.Sleep(10 * time.Millisecond)
				}
			}
		}(i)
	}

	// 4) Let it stabilize
	time.Sleep(5 * time.Second)

	// 5) Profile GC under steady load
	f, _ := os.Create("cpu.prof")
	pprof.StartCPUProfile(f)
	log.Println("Profiling for 30s under steady load…")
	time.Sleep(30 * time.Second)
	pprof.StopCPUProfile()
	f.Close()
	log.Println("Done, wrote cpu.prof")
}
