package disablesessions

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/prestonvasquez/dev/mongo-go-driver/v1/metrics"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/x/mongo/driver"
	"go.mongodb.org/mongo-driver/x/mongo/driver/session"
)

func TestDisablingSessions(t *testing.T) {
	const runDuration = 5 * time.Minute

	err := metrics.RunExp(func(ctx context.Context, coll *mongo.Collection) metrics.ExpResult {
		session.DisableSessionPooling = false
		query := bson.D{{Key: "field1", Value: "doesntexist"}}
		result := coll.FindOne(ctx, query)

		timeoutOps := 0
		if err := result.Err(); err != nil && errors.Is(err, context.DeadlineExceeded) {
			timeoutOps++
		}

		sessionIDSet := make(map[string]bool)
		driver.UniqueSessionIDs.Range(func(key, _ any) bool {
			sessionIDSet[key.(string)] = true
			driver.UniqueSessionIDs.Delete(key)

			return true
		})

		return metrics.ExpResult{
			OpCount:        1,
			TimeoutOpCount: timeoutOps,
			SessionIDSet:   sessionIDSet,
		}
	}, metrics.WithRunDuration(runDuration))

	require.NoError(t, err)
}

func TestDisablingSessionsMulti(t *testing.T) {
	const runDuration = 5 * time.Minute
	const preloadLargeCollection = 10_000

	err := metrics.RunExp(func(ctx context.Context, coll *mongo.Collection) metrics.ExpResult {
		session.DisableSessionPooling = true

		opsToAttempt := 10_000

		var timeoutOps atomic.Int32
		var ops atomic.Int32

		wg := sync.WaitGroup{}
		wg.Add(opsToAttempt)

		for i := 0; i < opsToAttempt; i++ {
			go func() {
				defer wg.Done()

				select {
				case <-ctx.Done():
					return
				default:
				}

				//query := bson.D{}

				query := bson.D{{Key: "field1", Value: "doesntexist"}}
				result := coll.FindOne(ctx, query)

				if err := result.Err(); err != nil && errors.Is(err, context.DeadlineExceeded) {
					timeoutOps.Add(1)
				}

				ops.Add(1)
			}()
		}

		wg.Wait()

		sessionIDSet := make(map[string]bool)
		driver.UniqueSessionIDs.Range(func(key, _ any) bool {
			sessionIDSet[key.(string)] = true
			driver.UniqueSessionIDs.Delete(key)

			return true
		})

		return metrics.ExpResult{
			OpCount:        int(ops.Load()),
			TimeoutOpCount: int(timeoutOps.Load()),
			SessionIDSet:   sessionIDSet,
		}
	},
		metrics.WithRunDuration(runDuration),
		metrics.WithPreloadCollectionSize(preloadLargeCollection))

	require.NoError(t, err)
}

//type experimentResult struct {
//	opCount int32
//}
//
//type experimentFn func(ctx context.Context, coll *mongo.Collection) (experimentResult, error)
//
//func BenchmarkSessionPooling(b *testing.B) {
//	// MongoDB connection URI
//	uri := os.Getenv("MONGODB_URI")
//	if uri == "" {
//		uri = "mongodb://localhost:27017"
//	}
//
//	// Connect to MongoDB
//	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(uri).SetMaxPoolSize(1))
//	if err != nil {
//		b.Fatalf("failed to connect to MongoDB: %v", err)
//	}
//
//	defer func() {
//		if err := client.Disconnect(context.Background()); err != nil {
//			b.Fatalf("Failed to disconnect MongoDB client: %v", err)
//		}
//	}()
//
//	db := client.Database("testdb")
//
//	collName, err := preloadLargeCollection(context.Background(), 100_000, client)
//	require.NoError(b, err)
//
//	coll := db.Collection(collName)
//
//	testCases := []struct {
//		name string
//		fn   experimentFn
//	}{
//		{
//			name: "single-threaded findOne",
//			fn:   singleThreadedFindOne,
//		},
//		{
//			name: "concurrent findOne",
//			fn:   multiThreadedFindOne,
//		},
//	}
//
//	for _, tt := range testCases {
//		for i := 0; i < 2; i++ {
//			session.DisableSessionPooling = i == 0
//
//			bname := fmt.Sprintf("%s DisableSessionPooling=%v", tt.name, session.DisableSessionPooling)
//			b.Run(bname, func(b *testing.B) {
//				b.ResetTimer()
//				b.ReportAllocs()
//
//				var opCount int32
//
//				b.RunParallel(func(pb *testing.PB) {
//					for pb.Next() {
//						//start := time.Now()
//						res, err := tt.fn(context.Background(), coll)
//						require.NoError(b, err)
//						atomic.AddInt32(&opCount, res.opCount)
//						//duration := time.Since(start)
//						//b.ReportMetric(float64(duration.Milliseconds()), "ms/op")
//					}
//				})
//
//				var uniqueSessionIDsLen int
//				driver.UniqueSessionIDs.Range(func(key, _ any) bool {
//					uniqueSessionIDsLen++
//					driver.UniqueSessionIDs.Delete(key)
//
//					return true
//				})
//
//				b.ReportMetric(float64(uniqueSessionIDsLen), "sessions")
//				//b.ReportMetric(float64(opCount), "concurrent_ops")
//			})
//		}
//	}
//}
//
//func singleThreadedFindOne(ctx context.Context, coll *mongo.Collection) (experimentResult, error) {
//	query := bson.D{
//		{Key: "field1", Value: "doesntexist"},
//
//		// By varyig the query we avoid potential for caching or query plan reuse
//		// that might mask the cost of creating new sessions.
//		{Key: "field1", Value: rand.Intn(1000)},
//	}
//
//	result := coll.FindOne(ctx, query)
//
//	err := result.Err()
//	if errors.Is(err, mongo.ErrNoDocuments) {
//		err = nil
//	}
//
//	return experimentResult{}, err
//}
//
//func multiThreadedFindOne(ctx context.Context, coll *mongo.Collection) (experimentResult, error) {
//	opsToAttempt := 1000
//
//	var opCount atomic.Int32
//
//	wg := sync.WaitGroup{}
//	wg.Add(opsToAttempt)
//
//	for i := 0; i < opsToAttempt; i++ {
//		go func() {
//			defer wg.Done()
//
//			select {
//			case <-ctx.Done():
//				return
//			default:
//			}
//
//			query := bson.D{
//				{Key: "field1", Value: "doesntexist"},
//				{Key: "field1", Value: rand.Intn(1000)},
//			}
//			coll.FindOne(ctx, query)
//
//			opCount.Add(1)
//		}()
//	}
//
//	wg.Wait()
//
//	return experimentResult{
//		opCount: opCount.Load(),
//	}, nil
//}
//
//// preloadLargeCollection populates a MongoDB collection with random data.
//func preloadLargeCollection(ctx context.Context, size int, client *mongo.Client) (string, error) {
//	collectionName := fmt.Sprintf("large_%s", uuid.NewString())
//	collection := client.Database("testdb").Collection(collectionName)
//
//	workerCount := runtime.NumCPU()
//	batchSize := size / workerCount
//
//	documents := make([]interface{}, batchSize)
//	for i := range documents {
//		documents[i] = bson.D{{Key: "field1", Value: rand.Int63()}, {Key: "field2", Value: rand.Int31()}}
//	}
//
//	errChan := make(chan error, workerCount)
//	doneChan := make(chan struct{}, workerCount)
//
//	for i := 0; i < workerCount; i++ {
//		go func() {
//			if _, err := collection.InsertMany(ctx, documents); err != nil {
//				errChan <- err
//			} else {
//				doneChan <- struct{}{}
//			}
//		}()
//	}
//
//	for i := 0; i < workerCount; i++ {
//		select {
//		case err := <-errChan:
//			return "", err
//		case <-doneChan:
//		}
//	}
//
//	return collectionName, nil
//}
