package metrics

import (
	"context"
	"fmt"
	"log"
	"os"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/exp/rand"
)

// Globals to manage worker states
var (
	workerCancelFuncsMu sync.Mutex
	workerCancelFuncs   []context.CancelFunc
	numWorkers          int32 // Atomic counter for the number of active workers
)

type ExpResult struct {
	OpCount        int
	TimeoutOpCount int
	SessionIDSet   map[string]bool
	Errors         map[string]int
}

type ExpFunc func(ctx context.Context, coll *mongo.Collection) ExpResult

// RunExp will run the given experiment function under the conditions provided
// as configurations.
func RunExp(experiment ExpFunc, cfgOpts ...ConfigOpt) error {
	cfg := NewConfig()
	for _, optFn := range cfgOpts {
		optFn(&cfg)
	}

	// MongoDB connection URI
	uri := os.Getenv("MONGODB_URI")
	if uri == "" {
		uri = "mongodb://localhost:27017"
	}

	// Connect to MongoDB
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(uri))
	if err != nil {
		return fmt.Errorf("failed to connect to MongoDB: %v", err)
	}

	defer func() {
		if err := client.Disconnect(context.Background()); err != nil {
			log.Fatalf("Failed to disconnect MongoDB client: %v", err)
		}
	}()

	// Preload data into a collection
	collName, err := preloadLargeCollection(context.Background(), 10000, client)
	if err != nil {
		return fmt.Errorf("failed to preload collection: %w", err)
	}

	db := client.Database("testdb")
	collection := db.Collection(collName)

	// Channels for communication
	latencyCh := make(chan time.Duration, 1000)
	startExpCh := make(chan struct{}, 1)

	// Start the latency aggregator
	go monitorLatency(cfg, latencyCh, collection, startExpCh)

	// Start the timeout experiment
	experimentContext, cancelExperiment := context.WithCancel(context.Background())

	go runExpAsync(experimentContext, collName, cfg, startExpCh, experiment)

	// Spawn initial workers
	spawnWorkers(cfg, cfg.initialWorkerCount, collection, latencyCh)

	time.Sleep(cfg.runDuration)
	cancelExperiment()

	// Clean up
	terminateAllWorkers()
	cancelExperiment()
	time.Sleep(10 * time.Second)

	return nil
}

func runExpAsync(ctx context.Context, collName string, cfg Config, signal <-chan struct{}, fn ExpFunc) {
	<-signal

	uri := os.Getenv("MONGODB_URI")
	if uri == "" {
		uri = "mongodb://localhost:27017"
	}

	poolMonitorCb := cfg.experimentPoolMonitorCallback
	if poolMonitorCb == nil {
		poolMonitorCb = newExpFuncPoolMonitor
	}

	poolMonitor := poolMonitorCb()

	commandMonitorCb := cfg.experimentCommandMonitorCallback
	if commandMonitorCb == nil {
		commandMonitorCb = newExpFuncCommandMonitor
	}

	commandMonitor := commandMonitorCb()

	client, err := mongo.Connect(context.Background(),
		options.Client().ApplyURI(uri).SetPoolMonitor(poolMonitor.PoolMonitor).
			SetMonitor(commandMonitor.CommandMonitor).SetTimeout(0), cfg.experimentClientOpts)
	if err != nil {
		log.Fatalf("failed to connect: %v", err)
	}

	defer func() {
		if err := client.Disconnect(context.Background()); err != nil {
			log.Fatalf("failed to disconnect: %v", err)
		}
	}()

	db := client.Database("metricsrunner")
	coll := db.Collection(collName)

	log.Println("[Experiment] starting timeout queries")

	opCount := 0
	timeoutErrCount := 0
	sessionIDSet := make(map[string]bool)
	var opDurs []float64

	errSet := make(map[string]int)

	for {
		select {
		case <-ctx.Done():
			type resultEntry struct {
				key   string
				value interface{}
			}

			setup := []resultEntry{
				{"target_latency", cfg.targetLatency},
				{"window_duration", cfg.windowDuration},
				{"initial_worker_count", cfg.initialWorkerCount},
				{"run_duration", cfg.runDuration},
				{"max_workers", cfg.maxWorkers},
			}

			if cfg.experimentTimeout != nil {
				setup = append(setup, resultEntry{"experiment_timeout", *cfg.experimentTimeout})
			}

			if cfg.experimentClientOpts != nil {
				setup = append(setup, resultEntry{"experiment_client_options", cfg.experimentClientOpts})
			}

			setup = append(setup, resultEntry{"preload_collection_size", cfg.preloadCollectionSize})

			log.Println("[Experiment] config:")
			for _, entry := range setup {
				log.Printf("  %s: %v", entry.key, entry.value)
			}

			results := []resultEntry{
				{"connections_closed", poolMonitor.ConnClosed.Load()},
				{"connections_closed_errors", poolMonitor.ConnClosedErrors},
				{"connections_closed_reasons", poolMonitor.ConnClosedReasons},
				{"connections_ready", len(poolMonitor.ConnReadyDur)},
				{"succeeded_pending_reads", poolMonitor.ConnPendingReadSucceeded.Load()},
				{"failed_pending_reads", poolMonitor.ConnPendingReadFailed.Load()},
				{"commands_failed", commandMonitor.Failed.Load()},
				{"commands_started", commandMonitor.Started.Load()},
				{"commands_succeeded", commandMonitor.Succeeded.Load()},
				{"average_connection_ready_duration_ms", average(poolMonitor.ConnReadyDur)},
				{"median_connection_ready_duration_ms", median(poolMonitor.ConnReadyDur)},
				{"op_count", opCount},
				{"timeout_err_count", timeoutErrCount},
				{"average_op_duration", average(opDurs)},
				{"median_op_duration", median(opDurs)},
				{"sessions", len(sessionIDSet)},
				{"errors", errSet},
			}
			log.Println("[Experiment] results:")
			for _, entry := range results {
				log.Printf("  %s: %v", entry.key, entry.value)
			}
			return
		default:
			expFnCtx, expFnCancel := context.WithCancel(context.Background())
			if cfg.experimentTimeout != nil {
				expFnCtx, expFnCancel = context.WithTimeout(ctx, *cfg.experimentTimeout)
			}

			opStart := time.Now()
			result := fn(expFnCtx, coll)
			opDurs = append(opDurs, float64(time.Since(opStart))/float64(time.Millisecond))

			opCount += result.OpCount
			timeoutErrCount += result.TimeoutOpCount

			if result.SessionIDSet != nil {
				for sessionID := range result.SessionIDSet {
					sessionIDSet[sessionID] = true
				}
			}

			for err, inc := range result.Errors {
				errSet[err] += inc
			}

			expFnCancel()
		}
	}
}

// median calculates the median of a sorted slice of float64 numbers.
func median(sortedData []float64) float64 {
	n := len(sortedData)
	if n == 0 {
		return 0 // Handle empty slice
	}
	if n%2 == 1 {
		return sortedData[n/2] // Odd number of elements
	}
	// Even number of elements
	mid := n / 2
	return (sortedData[mid-1] + sortedData[mid]) / 2
}

// average calculates the mean of a slice of float64 numbers.
func average(data []float64) float64 {
	if len(data) == 0 {
		return 0 // Handle empty slice to avoid division by zero
	}
	sum := 0.0
	for _, value := range data {
		sum += value
	}
	return sum / float64(len(data))
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
func spawnWorkers(cfg Config, count int, collection *mongo.Collection, latencyChannel chan<- time.Duration) {
	for i := 0; i < count; i++ {
		if atomic.LoadInt32(&numWorkers) >= cfg.maxWorkers {
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
func monitorLatency(
	cfg Config,
	latencyChannel chan time.Duration,
	collection *mongo.Collection,
	startExp chan<- struct{},
) {
	ticker := time.NewTicker(cfg.windowDuration)
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

			if averageMs < float64(cfg.targetLatency/time.Millisecond) {
				additionalWorkers := 5 * adjustmentWeight
				adjustmentWeight++
				log.Printf("[Monitor] Latency below target. Adding %d workers.", additionalWorkers)
				spawnWorkers(cfg, additionalWorkers, collection, latencyChannel)
			} else {
				log.Println("[Monitor] Latency above target. Triggering timeout queries.")
				startExp <- struct{}{}
			}

			lastAverage = average
		}
	}
}
