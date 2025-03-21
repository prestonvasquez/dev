package latency

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
	"go.mongodb.org/mongo-driver/event"
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

type experimentResult struct {
	ops        int
	timeoutOps int
	//map[string]interface{}
}

type experimentFn func(ctx context.Context, coll *mongo.Collection) experimentResult

func run(experiment experimentFn, cfgOpts ...configOpt) error {
	cfg := newConfig()
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

	db := client.Database("testdb")

	// Preload data into a collection
	collectionName, err := preloadLargeCollection(context.Background(), 10000, client)
	if err != nil {
		return fmt.Errorf("failed to preload collection: %w", err)
	}
	collection := db.Collection(collectionName)

	// Channels for communication
	latencyCh := make(chan time.Duration, 1000)
	startExpCh := make(chan struct{}, 1)

	// Start the latency aggregator
	go monitorLatency(cfg, latencyCh, collection, startExpCh)

	// Start the timeout experiment
	experimentContext, cancelExperiment := context.WithCancel(context.Background())
	go runExperiment(experimentContext, cfg, startExpCh, collectionName, experiment)

	// Spawn initial workers
	spawnWorkers(cfg, cfg.initialWorkerCount, collection, latencyCh)

	// Run for the specified duration
	time.Sleep(cfg.runDuration)

	// Clean up
	terminateAllWorkers()
	cancelExperiment()
	time.Sleep(10 * time.Second)

	return nil
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
func spawnWorkers(cfg config, count int, collection *mongo.Collection, latencyChannel chan<- time.Duration) {
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
	cfg config,
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

func runExperiment(ctx context.Context, cfg config, signal <-chan struct{}, collectionName string, fn experimentFn) {
	<-signal

	uri := os.Getenv("MONGODB_URI")
	if uri == "" {
		uri = "mongodb://localhost:27017"
	}

	//connectionPendingReadDurationMu := sync.Mutex{}
	//connectionPendingReadErrors := map[string]int{}
	//connectionPendingReadDurations := []float64{}

	var connectionPendingReadFailedCount atomic.Int32
	var connectionPendingReadSucceededCount atomic.Int32

	//topology.BGReadCallback = func(addr string, start, read time.Time, errs []error, connClosed bool) {
	//	connectionPendingReadDurationMu.Lock()
	//	defer connectionPendingReadDurationMu.Unlock()

	//	if !connClosed {
	//		connectionPendingReadSucceededCount++
	//	} else {
	//		connectionPendingReadFailedCount++
	//	}

	//	elapsed := time.Since(start)
	//	connectionPendingReadDurations = append(connectionPendingReadDurations, float64(elapsed.Milliseconds()))

	//	if len(errs) != 0 {
	//		for _, err := range errs {
	//			connectionPendingReadErrors[err.Error()]++
	//		}
	//	}
	//}

	var connectionsClosed atomic.Int64

	connectionReadyDurationsMu := sync.Mutex{}
	connectionReadyDurations := []float64{}

	connectionClosedMu := sync.Mutex{}
	connectionClosedErrors := map[string]int{}
	connectionClosedReasons := map[string]int{}

	poolMonitor := &event.PoolMonitor{
		Event: func(pe *event.PoolEvent) {
			switch pe.Type {
			case event.ConnectionPendingReadFailed:
				connectionPendingReadFailedCount.Add(1)
				//connectionPendingReadFailedCount.Add(1)

				//connectionPendingReadFailedReasonMu.Lock()
				//connectionPendingReadFailedReasons[pe.Reason] = struct{}{}
				//connectionPendingReadFailedReasonMu.Unlock()
			case event.ConnectionPendingReadSucceeded:
				connectionPendingReadSucceededCount.Add(1)

				//connectionPendingReadDurationMu.Lock()
				//connectionPendingReadDurations = append(connectionPendingReadDurations, float64(pe.Duration)/float64(time.Millisecond))
				//connectionPendingReadSucceededCount++
				//connectionPendingReadDurationMu.Unlock()
			case event.ConnectionClosed:
				connectionClosedMu.Lock()
				if pe.Error != nil {
					connectionClosedErrors[pe.Error.Error()]++
				}
				if pe.Reason != "" {
					connectionClosedReasons[pe.Reason]++
				}
				connectionClosedMu.Unlock()

				connectionsClosed.Add(1)
			case event.ConnectionReady:
				connectionReadyDurationsMu.Lock()
				connectionReadyDurations = append(connectionReadyDurations, float64(pe.Duration)/float64(time.Millisecond))
				connectionReadyDurationsMu.Unlock()
			}
		},
	}

	var commandFailed atomic.Int64
	var commandSucceeded atomic.Int64
	var commandStarted atomic.Int64

	cmdMonitor := &event.CommandMonitor{
		Started: func(_ context.Context, cse *event.CommandStartedEvent) {
			if cse.CommandName == "find" {
				commandStarted.Add(1)
			}
		},
		Succeeded: func(_ context.Context, cse *event.CommandSucceededEvent) {
			if cse.CommandName == "find" {
				commandSucceeded.Add(1)
			}
		},
		Failed: func(_ context.Context, evt *event.CommandFailedEvent) {
			if evt.CommandName == "find" {
				commandFailed.Add(1)
			}
		},
	}

	client, err := mongo.Connect(context.Background(),
		options.Client().ApplyURI(uri).SetPoolMonitor(poolMonitor).
			SetMonitor(cmdMonitor).SetTimeout(0), cfg.clientOpts)
	if err != nil {
		log.Fatalf("failed to connect: %v", err)
	}

	defer func() {
		if err := client.Disconnect(context.Background()); err != nil {
			log.Fatalf("failed to disconnect: %v", err)
		}
	}()

	db := client.Database("testdb")
	coll := db.Collection(collectionName)

	log.Println("[Experiment] starting timeout queries")

	opCount := 0
	timeoutErrCount := 0
	opDurs := []float64{}

	for {
		select {
		case <-ctx.Done():
			type resultEntry struct {
				key   string
				value interface{}
			}
			results := []resultEntry{
				{"connections_closed", connectionsClosed.Load()},
				{"connections_closed_errors", connectionClosedErrors},
				{"connections_closed_reasons", connectionClosedReasons},
				{"connections_ready", len(connectionReadyDurations)},
				{"succeeded_pending_reads", connectionPendingReadSucceededCount.Load()},
				{"failed_pending_reads", connectionPendingReadFailedCount.Load()},
				//{"pending_read_errors", connectionPendingReadErrors},
				{"commands_failed", commandFailed.Load()},
				{"commands_started", commandStarted.Load()},
				{"commands_succeeded", commandSucceeded.Load()},
				{"average_connection_ready_duration_ms", average(connectionReadyDurations)},
				{"median_connection_ready_duration_ms", median(connectionReadyDurations)},
				{"op_count", opCount},
				{"timeout_err_count", timeoutErrCount},
				//{"average_pending_read_dur_ms", average(connectionPendingReadDurations)},
				//{"median_pending_read_dur_ms", median(connectionPendingReadDurations)},
				{"average_op_duration", average(opDurs)},
				{"median_op_duration", median(opDurs)},
			}
			log.Println("[Experiment] results:")
			for _, entry := range results {
				log.Printf("  %s: %v", entry.key, entry.value)
			}
			return
		default:
			expFnCtx, expFnCancel := context.WithTimeout(ctx, cfg.experimentTimeout)
			expFnCtx = context.WithValue(expFnCtx, "latency_context", true)

			opStart := time.Now()
			result := fn(expFnCtx, coll)
			opDurs = append(opDurs, float64(time.Since(opStart))/float64(time.Millisecond))

			opCount += result.ops
			timeoutErrCount += result.timeoutOps

			expFnCancel()
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
