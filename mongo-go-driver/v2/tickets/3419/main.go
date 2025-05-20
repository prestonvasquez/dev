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

func main() {
	clientOpts := options.Client().
		SetMaxConnecting(500).
		SetMaxPoolSize(500) // Cap total open connections
	client, err := mongo.Connect(clientOpts)
	if err != nil {
		log.Fatalf("connect error: %v", err)
	}
	defer client.Disconnect(context.Background())

	// Spawn 500 goroutines pinging in a tight loop
	for i := 0; i < 500; i++ {
		go func() {
			for {
				if err := client.Ping(context.Background(), nil); err != nil {
					log.Printf("ping error: %v", err)
				}
			}
		}()
	}

	// Warm up: let the pool fill and a couple GC cycles run
	log.Println("warming up for 10s…")
	time.Sleep(10 * time.Second)

	// Start  CPU profile after warm-up
	f, err := os.Create("cpu.prof")
	if err != nil {
		log.Fatalf("could not create profile: %v", err)
	}
	pprof.StartCPUProfile(f)
	log.Println("profiling for 30s…")
	time.Sleep(30 * time.Second)
	pprof.StopCPUProfile()
	f.Close()
	log.Println("done profiling, wrote cpu.prof")
}
