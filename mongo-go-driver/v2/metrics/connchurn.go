package main

import (
	"context"
	"fmt"
	"log"
	"math"
	"os"
	"sort"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

const defaultConnectionChurnDB = "connectionChurnDB"
const throughputMetricsColl = "throughputMetrics"
const throughputMetricsStatsColl = "throughputMetricsStats"

// This operation will drop the metrics collection and run the analysis for the
// two latest runIDs. This is naive, but expedient.
//
// This also assume that the latest run is the baseline and the first run is the
// fix.

func main() {
	metricAtlasURI := os.Getenv("METRICS_ATLAS_URI")
	if metricAtlasURI == "" {
		log.Fatal("METRICS_ATLAS_URI required")
	}

	client, err := mongo.Connect(options.Client().ApplyURI(metricAtlasURI))
	if err != nil {
		log.Fatalf("failed to connect to metrics atlas database: %v", err)
	}

	type throughputMetric struct {
		RunID           string
		Baseline        float64
		Fix             float64
		BaselineSuccess float64
		FixSuccess      float64
		Percentile      float64
		Diff            float64
	}

	coll := client.Database(defaultConnectionChurnDB).Collection("throughput")

	latestTwoRunIDs, err := latestTwo(context.Background(), coll)
	if err != nil {
		log.Fatalf("failed to find latest two run ids: %v", err)
	}

	metricRunID := fmt.Sprintf("%d-%d", latestTwoRunIDs[0], latestTwoRunIDs[1])

	cur, err := coll.Find(context.Background(), bson.D{{"runid", bson.D{{"$in", latestTwoRunIDs}}}})
	if err != nil {
		log.Fatalf("failed to run find operation: %v", err)
	}

	type throughput struct {
		RunID             int64
		Percentile        float64
		ThroughputActual  float64
		ThroughputSuccess float64
	}

	trecords := []throughput{}
	if err := cur.All(context.Background(), &trecords); err != nil {
		log.Fatalf("failed to decode throughput data: %v", err)
	}

	metricsByPercentile := make(map[float64]*throughputMetric)

	diffs := make([]float64, 0, len(trecords))

	fixT := make([]float64, 0, len(trecords))
	basT := make([]float64, 0, len(trecords))

	fixTS := make([]float64, 0, len(trecords))
	basTS := make([]float64, 0, len(trecords))

	for _, rec := range trecords {
		mrecord := metricsByPercentile[rec.Percentile]
		if mrecord == nil {
			mrecord = &throughputMetric{
				RunID:      metricRunID,
				Percentile: rec.Percentile,
			}
		}

		if rec.RunID == latestTwoRunIDs[0] {
			mrecord.Fix = rec.ThroughputActual
			mrecord.FixSuccess = rec.ThroughputSuccess

			fixT = append(fixT, mrecord.Fix)
			fixTS = append(fixTS, mrecord.FixSuccess)
		}

		if rec.RunID == latestTwoRunIDs[1] {
			mrecord.Baseline = rec.ThroughputActual
			mrecord.BaselineSuccess = rec.ThroughputSuccess

			basT = append(basT, mrecord.Baseline)
			basTS = append(basTS, mrecord.BaselineSuccess)

			diffs = append(diffs, mrecord.Fix-mrecord.Baseline)
		}

		mrecord.Diff = mrecord.Fix - mrecord.Baseline

		metricsByPercentile[rec.Percentile] = mrecord
	}

	metrics := make([]throughputMetric, 0, len(metricsByPercentile))
	for _, rec := range metricsByPercentile {
		metrics = append(metrics, *rec)
	}

	coll = client.Database(defaultConnectionChurnDB).Collection(throughputMetricsColl)
	coll.Drop(context.Background())

	_, err = coll.InsertMany(context.Background(), metrics)
	if err != nil {
		log.Fatalf("failed to insert data: %v", err)
	}

	sort.Float64s(diffs)

	stats := struct {
		MedianDiff          float64
		MedianFix           float64
		MedianBaseline      float64
		MedianFixS          float64
		MedianBaselineS     float64
		AverageImprovement  float64
		AverageSImprovement float64
		AverageDiff         float64
		AverageFix          float64
		AverageBaseline     float64
		AverageFixS         float64
		AverageBaselineS    float64
		StdDiff             float64
	}{
		MedianDiff:          median(diffs),
		MedianFix:           median(fixT),
		MedianBaseline:      median(basT),
		MedianFixS:          median(fixTS),
		MedianBaselineS:     median(basTS),
		AverageImprovement:  ((average(fixT) - average(basT)) / average(basT)) * 100.0,
		AverageSImprovement: ((average(fixTS) - average(basTS)) / average(basTS)) * 100.0,
		AverageDiff:         average(diffs),
		AverageFix:          average(fixT),
		AverageBaseline:     average(basT),
		AverageFixS:         average(fixTS),
		AverageBaselineS:    average(basTS),
		StdDiff:             standardDeviation(diffs),
	}

	coll = client.Database(defaultConnectionChurnDB).Collection(throughputMetricsStatsColl)
	coll.Drop(context.Background())

	_, err = coll.InsertOne(context.Background(), stats)
	if err != nil {
		log.Fatalf("failed to insert stats: %v", err)
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

// standardDeviation calculates the standard deviation of a slice of float64 numbers.
func standardDeviation(data []float64) float64 {
	mean := 0.0
	for _, value := range data {
		mean += value
	}
	mean /= float64(len(data))
	var variance float64
	for _, value := range data {
		deviation := value - mean
		variance += deviation * deviation
	}
	variance /= float64(len(data))
	return math.Sqrt(variance)
}

// return the runID for the latest two runs for comparison.
func latestTwo(ctx context.Context, coll *mongo.Collection) ([]int64, error) {
	result := coll.Distinct(ctx, "runid", bson.D{})
	if err := result.Err(); err != nil {
		return nil, fmt.Errorf("failed to aggregate: %v", err)
	}

	ids := []int64{}
	if err := result.Decode(&ids); err != nil {
		return nil, fmt.Errorf("failed to decode results: %v", err)
	}

	// Sort the slice in ascending order
	sort.Slice(ids, func(i, j int) bool {
		return ids[i] < ids[j]
	})

	if len(ids) > 1 {
		return []int64{ids[len(ids)-2], ids[len(ids)-1]}, nil
	}

	return []int64{}, nil
}
