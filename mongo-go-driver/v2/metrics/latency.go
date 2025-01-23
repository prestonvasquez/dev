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
	"go.mongodb.org/mongo-driver/v2/event"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"golang.org/x/exp/rand"
)

var numWorkers = 20
var runDuration = 1 * time.Minute

const maxWorkers = 200
const windowDuration = 10 * time.Second
const targetLatency = 100 * time.Millisecond
const experimentTimeout = 50 * time.Millisecond

var workerCancelFuncsMu sync.Mutex
var workerCountMu sync.Mutex
var workerCancelFuncs = make([]context.CancelFunc, 0)

func main() {
	uri := os.Getenv("MONGODB_URI")
	if uri == "" {
		uri = "mongodb://localhost:27017"
	}

	client, err := mongo.Connect(options.Client().ApplyURI(uri))
	if err != nil {
		log.Fatalf("failed to connect: %v", err)
	}

	defer func() {
		if err := client.Disconnect(context.Background()); err != nil {
			log.Fatalf("failed to disconnect: %v", err)
		}
	}()

	db := client.Database("testdb")

	collName, err := loadLargeCollection(context.Background(), 10000, client)
	coll := db.Collection(collName)

	latencyChan := make(chan time.Duration, 1000)
	startTimeoutQueries := make(chan struct{}, 1)

	go aggregateLatency(latencyChan, coll, startTimeoutQueries)

	expCtx, cancelExp := context.WithCancel(context.Background())

	go runTimeoutQueries(expCtx, startTimeoutQueries, collName)

	spawnWorkers(numWorkers, coll, latencyChan)

	time.Sleep(runDuration)

	killAllWorkers()

	cancelExp()
	time.Sleep(10 * time.Second)
}

func killAllWorkers() {
	workerCancelFuncsMu.Lock()
	for _, cancelFunc := range workerCancelFuncs {
		cancelFunc()
	}

	workerCancelFuncs = nil
	workerCancelFuncsMu.Unlock()
}

func spawnWorkers(count int, coll *mongo.Collection, latencyChan chan<- time.Duration) {
	for i := 0; i < count; i++ {
		if numWorkers > maxWorkers {
			return
		}

		workerCountMu.Lock()
		numWorkers++
		workerCountMu.Unlock()

		ctx, cancel := context.WithCancel(context.Background())

		workerCancelFuncsMu.Lock()
		workerCancelFuncs = append(workerCancelFuncs, cancel)
		workerCancelFuncsMu.Unlock()

		go worker(ctx, coll, latencyChan)
	}
}

func worker(ctx context.Context, coll *mongo.Collection, latencyChan chan<- time.Duration) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			start := time.Now()

			query := bson.D{{Key: "field1", Value: "doesntexist"}}
			result := coll.FindOne(context.Background(), query)

			if err := result.Err(); err != nil && err != mongo.ErrNoDocuments {
				log.Printf("worker query error: %v", err)
			}

			end := time.Now()
			latency := end.Sub(start)

			latencyChan <- latency
		}
	}
}

func runTimeoutQueries(ctx context.Context, start <-chan struct{}, collName string) {
	<-start

	uri := os.Getenv("MONGODB_URI")
	if uri == "" {
		uri = "mongodb://localhost:27017"
	}

	var connectionsClosed atomic.Int64

	connectionReadyDurationsMu := sync.Mutex{}
	connectionReadyDurations := []float64{}

	connectionPendingReadDurationMu := sync.Mutex{}
	connectionPendingReadDurations := []float64{}
	connectionPendingCount := 0

	poolMonitor := &event.PoolMonitor{
		Event: func(pe *event.PoolEvent) {
			switch pe.Type {
			case event.ConnectionPendingReadDuration:
				connectionPendingReadDurationMu.Lock()
				connectionPendingReadDurations = append(connectionPendingReadDurations, float64(pe.Duration)/float64(time.Millisecond))
				connectionPendingCount++
				connectionPendingReadDurationMu.Unlock()
			case event.ConnectionClosed:
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

	client, err := mongo.Connect(options.Client().ApplyURI(uri).SetPoolMonitor(poolMonitor).SetMonitor(cmdMonitor))
	if err != nil {
		log.Fatalf("failed to connect: %v", err)
	}

	defer func() {
		if err := client.Disconnect(context.Background()); err != nil {
			log.Fatalf("failed to disconnect: %v", err)
		}
	}()

	db := client.Database("testdb")
	coll := db.Collection(collName)

	log.Println("[Experiment] starting timeout queries")

	opCount := 0
	timeoutErrCount := 0
	opDurs := []float64{}

	for {
		select {
		case <-ctx.Done():
			fmt.Println("DONE!", connectionReadyDurations)
			log.Printf(`[Experiment] results: {
	"connections_closed": %v,  
	"connections_ready": %v,
	"pending_reads": %v,
	"commands_failed": %v, 
	"commands_started": %v, 
	"commands_succeeded": %v,
	"average_connection_ready_duration_ms": %v, 
	"median_connection_ready_duration_ms": %v,
	"op_count": %v,
	"timeout_err_count": %v,
	"average_pending_read_dur_ms": %v,
	"median_pending_read_dur_ms": %v,
	"average_op_duration": %v, 
	"median_op_duration": %v
}`, connectionsClosed.Load(),
				len(connectionReadyDurations),
				connectionPendingCount,
				commandFailed.Load(),
				commandStarted.Load(),
				commandSucceeded.Load(),
				average(connectionReadyDurations),
				median(connectionReadyDurations),
				opCount,
				timeoutErrCount,
				average(connectionPendingReadDurations),
				median(connectionPendingReadDurations),
				average(opDurs),
				median(opDurs),
			)
			return
		default:
			ctx, cancel := context.WithTimeout(context.Background(), experimentTimeout)

			query := bson.D{{Key: "field1", Value: "doesntexist"}}

			opStart := time.Now()
			result := coll.FindOne(ctx, query)
			opDurs = append(opDurs, float64(time.Since(opStart))/float64(time.Millisecond))

			if errors.Is(result.Err(), context.DeadlineExceeded) {
				timeoutErrCount++
			}

			opCount++

			cancel()
		}
	}

}

func aggregateLatency(latencyChan chan time.Duration, coll *mongo.Collection, startTimeoutQueries chan<- struct{}) {
	ticker := time.NewTicker(windowDuration)
	defer ticker.Stop()

	var latencies []time.Duration
	var lastAvg float64 = -1

	spawnWeight := 1

	for {
		select {
		case latency, ok := <-latencyChan:
			if !ok {
				log.Println("latency channel closed.")

				return
			}

			latencies = append(latencies, latency)
		case <-ticker.C:
			if len(latencies) == 0 {
				log.Println("no latencies recorded in this window")
				continue
			}

			var sum time.Duration
			for _, l := range latencies {
				sum += l
			}

			avg := float64(sum) / float64(len(latencies))

			trend := "decreaing or stable"
			if avg > lastAvg {
				trend = "increasing"
			}

			avgMs := avg / float64(time.Millisecond)
			log.Printf("[Latency Aggregator] Average latency for last %v: %.2f ms (%s)\n",
				windowDuration, avgMs, trend)

			if avgMs < float64(targetLatency)/float64(time.Millisecond) {
				additional := 5 * spawnWeight
				spawnWeight++

				log.Printf("[Aggregator] Latency stable or decreasing; spawning %d more workers...\n", additional)
				spawnWorkers(additional, coll, latencyChan)
			} else {
				startTimeoutQueries <- struct{}{}
			}

			latencies = []time.Duration{}
			lastAvg = avg
		}
	}
}

// loadLargeCollection will dedicate a worker pool to inserting test data into
// an unindexed collection. Each record is 31 bytes in size.
func loadLargeCollection(ctx context.Context, size int, client *mongo.Client) (string, error) {
	// Initialize a collection with the name "large<uuid>".
	collName := fmt.Sprintf("large%s", uuid.NewString())

	goRoutines := runtime.NumCPU()

	// Partition the volume into equal sizes per go routine. Use the floor if the
	// volume is not divisible by the number of goroutines.
	perGoroutine := size / goRoutines

	docs := make([]interface{}, perGoroutine)
	for i := range docs {
		docs[i] = bson.D{
			{Key: "field1", Value: rand.Int63()},
			{Key: "field2", Value: rand.Int31()},
		}
	}

	errs := make(chan error, goRoutines)
	done := make(chan struct{}, goRoutines)

	coll := client.Database("testdb").Collection(collName)

	for i := 0; i < int(goRoutines); i++ {
		go func(i int) {
			_, err := coll.InsertMany(ctx, docs)
			if err != nil {
				errs <- fmt.Errorf("goroutine %v failed: %w", i, err)
			}

			done <- struct{}{}
		}(i)
	}

	go func() {
		defer close(errs)
		defer close(done)

		for i := 0; i < int(goRoutines); i++ {
			<-done
		}
	}()

	// Await errors and return the first error encountered.
	for err := range errs {
		if err != nil {
			return "", err
		}
	}

	return collName, nil
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
