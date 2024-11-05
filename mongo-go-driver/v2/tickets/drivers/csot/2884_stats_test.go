package csot

import (
	"context"
	"fmt"
	"log"
	"math"
	"os"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"math/rand"

	"github.com/google/uuid"
	"github.com/jedib0t/go-pretty/table"
	"github.com/montanaflynn/stats"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/event"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"gonum.org/v1/gonum/integrate"
)

const defaultURI = "mongodb://localhost:27017"

type csotCommand struct {
	config string // name of test configuartion in the /test directory.
}

type testCase struct {
	Description string `yaml:"description"` // Name of the test
	Volume      uint   `yaml:"volume"`      // Number of records to load
	GoRoutines  uint   `yaml:"goRoutines"`  // Number of goroutines to evenly split op execution
	MaxPoolSize uint64 `yaml:"maxPoolSize"`

	// rttPercentile is the rtt percentile to use as the timeout for the
	// short-circuit operation used for the text case. The valid interval for
	// this field is [0,1]. If set to "0", then the minimum RTT is used. If set
	// to "1", then the maximum RTT is used.
	rttPercentile float64
}

type test struct {
	Cases      []testCase `yaml:"cases"`      // Test cases
	SampleSize int        `yaml:"sampleSize"` // Number of times to run test cases
}

// loadLargeCollection bulk-insersects data into a collection for the test
// case.
func loadLargeCollection(t *testing.T, coll *mongo.Collection, tcase testCase) {
	t.Helper()

	t.Logf("loading large collection for %q", tcase.Description)
	defer t.Logf("finished loading large collection for %q", tcase.Description)

	const batchSize = 500 // Size of batches to load for testing

	docs := make([]interface{}, 0, batchSize)
	for i := 0; i < batchSize; i++ {
		docs = append(docs, bson.D{
			{"field1", rand.Int63()},
			{"field2", rand.Int31()},
		})
	}

	// Partition the volume into equal sizes per go routine. Use the floor if the
	// volume is not divisible by the number of goroutines.
	perGoroutine := tcase.Volume / tcase.GoRoutines

	// Number of batches to insert per goroutine. Use the floor if perGoroutine
	// is not divisible by the batchSize.
	batches := perGoroutine / batchSize

	errs := make(chan error, tcase.GoRoutines)
	done := make(chan struct{}, tcase.GoRoutines)

	for i := 0; i < int(tcase.GoRoutines); i++ {
		go func(i int) {
			for j := 0; j < int(batches); j++ {
				_, err := coll.InsertMany(context.Background(), docs)
				if err != nil {
					errs <- fmt.Errorf("goroutine %v failed: %w", i, err)

					break
				}
			}

			done <- struct{}{}
		}(i)
	}

	go func() {
		defer close(errs)

		for i := 0; i < int(tcase.GoRoutines); i++ {
			<-done
		}
	}()

	// Await errors and return the first error encountered.
	for err := range errs {
		require.NoError(t, err)
	}
}

type latencyStats struct {
	max        time.Duration
	min        time.Duration
	median     time.Duration
	mean       time.Duration
	percentile time.Duration
}

func getStats(t *testing.T, times []time.Duration, tcase testCase) *latencyStats {
	t.Helper()

	samples := make(stats.Float64Data, len(times))
	for i := range times {
		samples[i] = float64(times[i])
	}

	maxv, err := stats.Max(samples)
	require.NoError(t, err, "failed to get max from samples")

	minv, err := stats.Min(samples)
	require.NoError(t, err, "failed to get min from samples")

	medv, err := stats.Median(samples)
	require.NoError(t, err, "failed to get median from samples")

	mean, err := stats.Mean(samples)
	require.NoError(t, err, "failed to get mean from samples")

	var percentile float64
	switch tcase.rttPercentile {
	case 0.0:
		percentile = minv
	case 1.0:
		percentile = maxv
	default:
		percentile, err = stats.Percentile(samples, tcase.rttPercentile*100)
		require.NoError(t, err, "failed to cacluate percentile %v", tcase.rttPercentile*100)
	}

	return &latencyStats{
		max:        time.Duration(maxv),
		min:        time.Duration(minv),
		median:     time.Duration(medv),
		mean:       time.Duration(mean),
		percentile: time.Duration(percentile),
	}
}

func getQueryStats(t *testing.T, coll *mongo.Collection, query bson.D, tcase testCase) *latencyStats {
	samples := make([]time.Duration, 0, tcase.GoRoutines*10)
	var samplesMu sync.Mutex

	errs := make(chan error, tcase.GoRoutines)
	done := make(chan struct{}, tcase.GoRoutines)

	for i := 0; i < int(tcase.GoRoutines); i++ {
		go func() {
			// Higher durations yield more accurate statistics.
			durations := make([]time.Duration, 10)
			for i := 0; i < len(durations); i++ {
				start := time.Now()
				err := coll.FindOne(context.Background(), query).Err()
				durations[i] = time.Since(start)

				if err != nil && err != mongo.ErrNoDocuments {
					errs <- fmt.Errorf("failed to collect query stats: %w", err)

					break
				}
			}

			samplesMu.Lock()
			samples = append(samples, durations...)
			samplesMu.Unlock()

			done <- struct{}{}
		}()
	}

	go func() {
		defer close(errs)

		for i := 0; i < int(tcase.GoRoutines); i++ {
			<-done
		}
	}()

	for err := range errs {
		require.NoError(t, err)
	}

	samplesMu.Lock()
	defer samplesMu.Unlock()

	return getStats(t, samples[:], tcase)
}

type result struct {
	Description       string
	Percentile        float64
	ShortCircuitRate  float64
	timeoutErrCount   int
	failRate          float64
	successRate       float64
	connectionsClosed int64

	bgReadErrsCount          int64
	bgConnectionsClosedCount int64
	bgRoutinesCount          int64
}

func runTestCase(t *testing.T, tcase testCase) *result {
	t.Helper()

	uri := os.Getenv("MONGODB_URI")
	if uri == "" {
		uri = defaultURI
	}

	var connectionsClosed int64
	poolMonitor := &event.PoolMonitor{
		Event: func(pe *event.PoolEvent) {
			if pe.Type == event.ConnectionClosed {
				atomic.AddInt64(&connectionsClosed, 1)
			}
		},
	}

	var commandFailed int64
	var commandSucceeded int64
	var commandStarted int64
	cmdMonitor := &event.CommandMonitor{
		Started: func(_ context.Context, cse *event.CommandStartedEvent) {
			if cse.CommandName == "find" {
				log.Printf("started: %+v\n", cse)
				atomic.AddInt64(&commandStarted, 1)
			}
		},
		Succeeded: func(_ context.Context, cse *event.CommandSucceededEvent) {
			if cse.CommandName == "find" {
				log.Printf("succeeded: %+v\n", cse)
				atomic.AddInt64(&commandSucceeded, 1)
			}
		},
		Failed: func(_ context.Context, evt *event.CommandFailedEvent) {
			if evt.CommandName == "find" {
				log.Printf("failed: %+v\n", evt)
				atomic.AddInt64(&commandFailed, 1)
			}
		},
	}

	clientOpts := options.Client().SetTimeout(0).ApplyURI(uri).
		SetPoolMonitor(poolMonitor).SetMonitor(cmdMonitor)

	client, err := mongo.Connect(clientOpts)
	require.NoError(t, err)

	defer func() {
		_ = client.Disconnect(context.Background())
	}()

	// Initialize a collection with the name "large<uuid>".
	uuid := uuid.New()
	collName := fmt.Sprintf("large%x%x%x%x%x", uuid[0:4], uuid[4:6], uuid[6:8], uuid[8:10], uuid[10:])

	unindexedColl := client.Database("testdb").Collection(collName)
	defer func() { unindexedColl.Drop(context.Background()) }()

	// Load the test data.
	loadLargeCollection(t, unindexedColl, tcase)

	// Use the query that will be used in the "find" operations to get a various
	// query statistics that can be used to determine how to timeout an operation.
	query := bson.D{{"field1", "doesntexist"}}

	qstats := getQueryStats(t, unindexedColl, query, tcase)

	timeouts := make([]time.Duration, tcase.Volume)
	for i := 0; i < int(tcase.Volume); i++ {
		timeouts[i] = qstats.percentile
	}

	// Partition the volume into equal sizes per go routine. Use the floor if the
	// volume is not divisible by the number of goroutines.
	perGoroutine := tcase.Volume / tcase.GoRoutines

	errs := make(chan error, tcase.GoRoutines*perGoroutine)
	done := make(chan struct{}, tcase.GoRoutines)

	// Run the find query on an unindex collection in partitions upto the number
	// of goroutines.
	for i := 0; i < int(tcase.GoRoutines); i++ {
		go func(i int) {
			for j := 0; j < int(perGoroutine); j++ {
				idx := (int(perGoroutine)-1)*i + j + i
				timeout := timeouts[idx]

				ctx, cancel := context.WithTimeout(context.Background(), timeout)

				err := unindexedColl.FindOne(ctx, query).Err()
				cancel()

				if err != nil && err != mongo.ErrNoDocuments {
					errs <- err

					fmt.Println("given error by op")
				}
			}

			fmt.Println("done", i)
			done <- struct{}{}
		}(i)
	}

	go func() {
		defer close(errs)
		defer fmt.Println("errs closed")
		for i := 0; i < int(tcase.GoRoutines); i++ {
			<-done
		}
	}()

	fmt.Println("before err")
	gotTimeoutErrCount := 0
	for err := range errs {
		if mongo.IsTimeout(err) {
			// We don't consider "ErrDeadlineWouldBeExceeded" errors, these would not
			// result in a connection closing.
			gotTimeoutErrCount++
		}
	}

	fmt.Println("after err")

	shortCircuitRate := 1.0
	if gotTimeoutErrCount != 0 {
		shortCircuitRate = 1.0 - float64(connectionsClosed)/float64(gotTimeoutErrCount)
	}

	fmt.Println("after scr")

	// Need to subtract 100 for the query stats
	failRate := float64(commandFailed) / float64(tcase.Volume)
	successRate := float64(commandSucceeded-100.0) / float64(tcase.Volume)

	fmt.Println(failRate, successRate)

	return &result{
		Description:       tcase.Description,
		ShortCircuitRate:  shortCircuitRate,
		Percentile:        tcase.rttPercentile,
		timeoutErrCount:   gotTimeoutErrCount,
		failRate:          failRate,
		connectionsClosed: connectionsClosed,
		successRate:       successRate,
	}
}

func runTestCasePercentiles(t *testing.T, tcase testCase) []*result {
	t.Helper()

	if tcase.rttPercentile > 1.0 || tcase.rttPercentile < 0.0 {
		t.Errorf("rtt percentile must be in [0,1]")
	}

	results := make([]*result, 101)
	for p := 0; p <= 100; p++ {
		tcase := tcase
		tcase.rttPercentile = float64(p) / 100.0

		t.Logf("running test case %q for percentile %v", tcase.Description, p)
		exResult := runTestCase(t, tcase)

		t.Logf("finished running test case %q for percentile %v", tcase.Description, p)

		results[p] = exResult
	}

	return results
}

// calculateSimpsonsE returns an approximation of the area under the curve of
// discrete data defined by [100 percentiles, short-circuit succeeded].
func calculateSimpsonsE(results []*result) float64 {
	ff64s := []float64{}
	NaNSet := map[int]bool{}

	for idx, result := range results {
		if !math.IsNaN(result.ShortCircuitRate) {
			ff64s = append(ff64s, result.ShortCircuitRate)
		} else {
			NaNSet[idx] = true
		}
	}

	fx64s := []float64{}
	for idx, result := range results {
		if NaNSet[idx] {
			continue
		}

		fx64s = append(fx64s, result.Percentile)
	}

	return integrate.Simpsons(fx64s, ff64s)
}

// calcFailRate returns the rate at which the tests resulting in a
// successful round trip to the server.
func calcFailRate(results []*result) float64 {
	var sum float64
	for _, res := range results {
		sum += res.failRate
	}

	return sum / float64(len(results))
}

func calcSuccessRate(results []*result) float64 {
	var sum float64
	for _, res := range results {
		sum += res.successRate
	}

	return sum / float64(len(results))
}

func calcConnectionsClosedRate(results []*result) float64 {
	var sum float64
	for _, res := range results {
		sum += float64(res.connectionsClosed)
	}

	return sum / float64(len(results))
}

func calcConnectionsClosedMedian(results []*result) float64 {
	cc := make([]float64, len(results))
	for idx, res := range results {
		cc[idx] = float64(res.connectionsClosed)
	}

	med, _ := stats.Median(cc)

	return med
}

func calcTimeoutErrRate(results []*result) float64 {
	var sum float64
	for _, res := range results {
		sum += float64(res.timeoutErrCount)
	}

	return sum / float64(len(results))
}

func calcTimeoutErrMedian(results []*result) float64 {
	cc := make([]float64, len(results))
	for idx, res := range results {
		cc[idx] = float64(res.timeoutErrCount)
	}

	med, _ := stats.Median(cc)

	return med
}

func runTestSample(t *testing.T, tbl table.Writer, cfg test) {
	for _, tcase := range cfg.Cases {
		start := time.Now()
		results := runTestCasePercentiles(t, tcase)

		end := time.Since(start)

		tbl.AppendRow(table.Row{
			tcase.Description,
			tcase.Volume,
			end,
			//fmt.Sprintf("%.4f", calcFailRate(results)),
			fmt.Sprintf("%.4f", calcSuccessRate(results)),
			fmt.Sprintf("%.4f", calcConnectionsClosedRate(results)),
			//fmt.Sprintf("%.4f", calcConnectionsClosedMedian(results)),
			//fmt.Sprintf("%.4f", calcTimeoutErrRate(results)),
			//fmt.Sprintf("%.4f", calcTimeoutErrMedian(results)),
			//fmt.Sprintf("%.4f", rMean),
			//fmt.Sprintf("%.4f", cMean),
			//fmt.Sprintf("%.4f", eMean),
			fmt.Sprintf("%.4f", calculateSimpsonsE(results)),
			fmt.Sprintf("%.4f", 1.0-calculateSimpsonsE(results)),
		})
	}

	return
}

func runTest(t *testing.T, tbl table.Writer, cfg test) {
	for i := 0; i < cfg.SampleSize; i++ {
		runTestSample(t, tbl, cfg)
	}
}

func TestConnectionChurn(t *testing.T) {
	test := test{
		SampleSize: 1,
		Cases: []testCase{
			{
				Description: "low volume",
				GoRoutines:  10,
				Volume:      50,
				MaxPoolSize: 1,
			},
		},
	}

	tbl := table.NewWriter()
	tbl.SetOutputMirror(os.Stdout)

	tbl.AppendHeader(table.Row{
		"test",
		"volume",
		"elapsed",
		"success rate",
		"conn. closed mean",
		"simpsons (s)",
		"1 - s",
	})

	runTest(t, tbl, test)

	tbl.Render()
}
