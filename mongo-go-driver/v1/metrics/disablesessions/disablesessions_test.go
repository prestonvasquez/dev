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

		errSet := make(map[error]int)
		errSetMu := sync.Mutex{}

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

				err := result.Err()
				if err != nil && errors.Is(err, context.DeadlineExceeded) {
					timeoutOps.Add(1)
				}

				if err != nil {
					errSetMu.Lock()
					errSet[err]++
					errSetMu.Unlock()
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
			Errors:         errSet,
		}
	},
		metrics.WithRunDuration(runDuration),
		metrics.WithPreloadCollectionSize(preloadLargeCollection))

	require.NoError(t, err)
}
