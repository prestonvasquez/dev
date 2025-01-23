// MongoDB Latency Testing and Monitoring Program
package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"golang.org/x/exp/rand"
)

// Configuration constants
const (
	targetLatency      = 100 * time.Millisecond // Desired latency target
	windowDuration     = 10 * time.Second       // Time window for aggregating latency
	runDuration        = 1 * time.Minute        // Duration to run the test
	maxWorkers         = 200                    // Maximum number of workers allowed
	experimentTimeout  = 50 * time.Millisecond  // Timeout for experiment queries
	initialWorkerCount = 20                     // Initial number of workers
)

// Globals to manage worker states
var (
	workerCancelFuncsMu sync.Mutex
	workerCancelFuncs   []context.CancelFunc
	numWorkers          int32 // Atomic counter for the number of active workers
)

func main() {
	// MongoDB connection URI
	uri := os.Getenv("MONGODB_URI")
	if uri == "" {
		uri = "mongodb://localhost:27017"
	}

	// Connect to MongoDB
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(uri))
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer func() {
		if err := client.Disconnect(context.Background()); err != nil {
			log.Fatalf("Failed to disconnect MongoDB client: %v", err)
		}
	}()

	db := client.Database("testdb")

	// Preload data into a collection
	collectionName, err := preloadLargeCollection(context.Background(), 10000, client)
	if err != nil {
		log.Fatalf("Failed to preload collection: %v", err)
	}
	collection := db.Collection(collectionName)

	// Channels for communication
	latencyChannel := make(chan time.Duration, 1000)
	timeoutSignal := make(chan struct{}, 1)

	// Start the latency aggregator
	go monitorLatency(latencyChannel, collection, timeoutSignal)

	// Start the timeout experiment
	experimentContext, cancelExperiment := context.WithCancel(context.Background())
	go runTimeoutExperiment(experimentContext, timeoutSignal, collectionName)

	// Spawn initial workers
	spawnWorkers(initialWorkerCount, collection, latencyChannel)

	// Run for the specified duration
	time.Sleep(runDuration)

	// Clean up
	terminateAllWorkers()
	cancelExperiment()
	time.Sleep(10 * time.Second)
}

// terminateAllWorkers stops all active workers by calling their cancel functions.
func terminateAllWorkers() {
	workerCancelFuncsMu.Lock()
	defer workerCancelFuncsMu.Unlock()
	for _, cancelFunc := range workerCancelFuncs {
		cancelFunc()
	}
	workerCancelFuncs = nil
}

// spawnWorkers starts a specified number of workers.
func spawnWorkers(count int, collection *mongo.Collection, latencyChannel chan<- time.Duration) {
	for i := 0; i < count; i++ {
		if atomic.LoadInt32(&numWorkers) >= maxWorkers {
			return
		}

		ctx, cancel := context.WithCancel(context.Background())

		workerCancelFuncsMu.Lock()
		workerCancelFuncs = append(workerCancelFuncs, cancel)
		workerCancelFuncsMu.Unlock()

		atomic.AddInt32(&numWorkers, 1)
		go worker(ctx, collection, latencyChannel)
	}
}

// worker performs MongoDB queries and sends latency data to the latency channel.
func worker(ctx context.Context, collection *mongo.Collection, latencyChannel chan<- time.Duration) {
	defer atomic.AddInt32(&numWorkers, -1)
	for {
		select {
		case <-ctx.Done():
			return
		default:
			start := time.Now()

			query := bson.D{{Key: "field1", Value: "doesntexist"}}
			result := collection.FindOne(context.Background(), query)

			if err := result.Err(); err != nil && err != mongo.ErrNoDocuments {
				log.Printf("Worker query error: %v", err)
			}

			latency := time.Since(start)
			latencyChannel <- latency
		}
	}
}

// monitorLatency aggregates latencies and adjusts workers dynamically based on trends.
func monitorLatency(latencyChannel chan time.Duration, collection *mongo.Collection, timeoutSignal chan<- struct{}) {
	ticker := time.NewTicker(windowDuration)
	defer ticker.Stop()

	var latencies []time.Duration
	var lastAverage float64 = -1
	adjustmentWeight := 1

	for {
		select {
		case latency, ok := <-latencyChannel:
			if !ok {
				log.Println("Latency channel closed.")
				return
			}
			latencies = append(latencies, latency)
		case <-ticker.C:
			if len(latencies) == 0 {
				log.Println("No latencies recorded in this window.")
				continue
			}

			sum := time.Duration(0)
			for _, latency := range latencies {
				sum += latency
			}
			average := float64(sum) / float64(len(latencies))
			latencies = nil

			averageMs := average / float64(time.Millisecond)
			trend := "stable or decreasing"
			if average > lastAverage {
				trend = "increasing"
			}

			log.Printf("[Monitor] Average latency: %.2f ms (%s)", averageMs, trend)

			if averageMs < float64(targetLatency/time.Millisecond) {
				additionalWorkers := 5 * adjustmentWeight
				adjustmentWeight++
				log.Printf("[Monitor] Latency below target. Adding %d workers.", additionalWorkers)
				spawnWorkers(additionalWorkers, collection, latencyChannel)
			} else {
				log.Println("[Monitor] Latency above target. Triggering timeout queries.")
				timeoutSignal <- struct{}{}
			}

			lastAverage = average
		}
	}
}

// runTimeoutExperiment executes queries with a timeout to simulate stress on MongoDB.
func runTimeoutExperiment(ctx context.Context, signal <-chan struct{}, collectionName string) {
	<-signal

	uri := os.Getenv("MONGODB_URI")
	if uri == "" {
		uri = "mongodb://localhost:27017"
	}

	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(uri))
	if err != nil {
		log.Fatalf("Failed to connect for timeout experiment: %v", err)
	}
	defer func() {
		if err := client.Disconnect(context.Background()); err != nil {
			log.Fatalf("Failed to disconnect timeout experiment client: %v", err)
		}
	}()

	collection := client.Database("testdb").Collection(collectionName)
	log.Println("[Timeout Experiment] Starting queries with timeouts.")

	opCount := 0
	timeoutCount := 0

	for {
		select {
		case <-ctx.Done():
			log.Printf("[Timeout Experiment] Completed. Total operations: %d, Timeouts: %d", opCount, timeoutCount)
			return
		default:
			queryCtx, cancel := context.WithTimeout(context.Background(), experimentTimeout)
			query := bson.D{{Key: "field1", Value: "doesntexist"}}

			start := time.Now()
			result := collection.FindOne(queryCtx, query)
			cancel()

			if errors.Is(result.Err(), context.DeadlineExceeded) {
				timeoutCount++
			}
			opCount++
		}
	}
}

// preloadLargeCollection populates a MongoDB collection with random data.
func preloadLargeCollection(ctx context.Context, size int, client *mongo.Client) (string, error) {
	collectionName := fmt.Sprintf("large_%s", uuid.NewString())
	collection := client.Database("testdb").Collection(collectionName)

	workerCount := runtime.NumCPU()
	batchSize := size / workerCount

	documents := make([]interface{}, batchSize)
	for i := range documents {
		documents[i] = bson.D{{Key: "field1", Value: rand.Int63()}, {Key: "field2", Value: rand.Int31()}}
	}

	errChan := make(chan error, workerCount)
	doneChan := make(chan struct{}, workerCount)

	for i := 0; i < workerCount; i++ {
		go func() {
			if _, err := collection.InsertMany(ctx, documents); err != nil {
				errChan <- err
			} else {
				doneChan <- struct{}{}
			}
		}()
	}

	for i := 0; i < workerCount; i++ {
		select {
		case err := <-errChan:
			return "", err
		case <-doneChan:
		}
	}

	return collectionName, nil
}
