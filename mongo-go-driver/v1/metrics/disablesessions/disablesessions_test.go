package disablesessions

import (
	"context"
	"errors"
	"fmt"
	"os"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/x/mongo/driver"
	"go.mongodb.org/mongo-driver/x/mongo/driver/session"
	"golang.org/x/exp/rand"
)

type experimentResult struct {
	opCount int32
}

type experimentFn func(ctx context.Context, coll *mongo.Collection) (experimentResult, error)

func BenchmarkSessionPooling(b *testing.B) {
	// MongoDB connection URI
	uri := os.Getenv("MONGODB_URI")
	if uri == "" {
		uri = "mongodb://localhost:27017"
	}

	// Connect to MongoDB
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(uri))
	if err != nil {
		b.Fatalf("failed to connect to MongoDB: %v", err)
	}

	defer func() {
		if err := client.Disconnect(context.Background()); err != nil {
			b.Fatalf("Failed to disconnect MongoDB client: %v", err)
		}
	}()

	db := client.Database("testdb")

	collName, err := preloadLargeCollection(context.Background(), 100_000, client)
	require.NoError(b, err)

	coll := db.Collection(collName)

	testCases := []struct {
		name string
		fn   experimentFn
	}{
		{
			name: "single-threaded findOne",
			fn:   singleThreadedFindOne,
		},
		{
			name: "concurrent findOne",
			fn:   multiThreadedFindOne,
		},
	}

	for _, tt := range testCases {
		for i := 0; i < 2; i++ {
			session.DisableSessionPooling = i == 0

			b.Run(tt.name+fmt.Sprintf(" DisableSessionPooling=%v", session.DisableSessionPooling), func(b *testing.B) {
				b.ResetTimer()
				b.ReportAllocs()

				var opCount int32

				for i := 0; i < b.N; i++ {
					res, err := tt.fn(context.Background(), coll)
					require.NoError(b, err)

					opCount += res.opCount
				}

				var uniqueSessionIDsLen int
				driver.UniqueSessionIDs.Range(func(key, _ any) bool {
					uniqueSessionIDsLen++

					driver.UniqueSessionIDs.Delete(key)

					return true
				})

				b.ReportMetric(float64(uniqueSessionIDsLen), "sessions")
				b.ReportMetric(float64(opCount), "concurrent_ops")
			})
		}
	}
}

func singleThreadedFindOne(ctx context.Context, coll *mongo.Collection) (experimentResult, error) {
	query := bson.D{{Key: "field1", Value: "doesntexist"}}
	result := coll.FindOne(ctx, query)

	err := result.Err()
	if errors.Is(err, mongo.ErrNoDocuments) {
		err = nil
	}

	return experimentResult{}, err
}

func multiThreadedFindOne(ctx context.Context, coll *mongo.Collection) (experimentResult, error) {
	opsToAttempt := 100

	var opCount atomic.Int32

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

			query := bson.D{{Key: "field1", Value: "doesntexist"}}
			coll.FindOne(ctx, query)

			opCount.Add(1)
		}()
	}

	wg.Wait()

	return experimentResult{
		opCount: opCount.Load(),
	}, nil
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
