package latency

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func TestLatency(t *testing.T) {
	t.Run("single-threaded findOne", func(t *testing.T) {
		err := run(func(ctx context.Context, coll *mongo.Collection) experimentResult {
			query := bson.D{{Key: "field1", Value: "doesntexist"}}
			result := coll.FindOne(ctx, query)

			timeoutOps := 0
			if err := result.Err(); err != nil && errors.Is(err, context.DeadlineExceeded) {
				timeoutOps++
			}

			return experimentResult{
				ops:        1,
				timeoutOps: timeoutOps,
			}
		})
		require.NoError(t, err, "failed to run experiment")
	})

	t.Run("multi-threaded findOne", func(t *testing.T) {
		err := run(func(ctx context.Context, coll *mongo.Collection) experimentResult {
			opsToAttempt := 100

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

					query := bson.D{{Key: "field1", Value: "doesntexist"}}
					result := coll.FindOne(ctx, query)

					if err := result.Err(); err != nil && errors.Is(err, context.DeadlineExceeded) {
						timeoutOps.Add(1)
					}

					ops.Add(1)
				}()
			}

			wg.Wait()

			return experimentResult{
				ops:        int(ops.Load()),
				timeoutOps: int(timeoutOps.Load()),
			}
		})
		require.NoError(t, err, "failed to run experiment")
	})

	t.Run("multi-threaded findOne with maxPoolSize=1", func(t *testing.T) {
		clientOpts := options.Client().SetMaxPoolSize(1)

		err := run(func(ctx context.Context, coll *mongo.Collection) experimentResult {
			opsToAttempt := 100

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

					query := bson.D{{Key: "field1", Value: "doesntexist"}}
					result := coll.FindOne(ctx, query)

					if err := result.Err(); err != nil && errors.Is(err, context.DeadlineExceeded) {
						timeoutOps.Add(1)
					}

					ops.Add(1)
				}()
			}

			wg.Wait()

			return experimentResult{
				ops:        int(ops.Load()),
				timeoutOps: int(timeoutOps.Load()),
			}
		}, withClientOptions(clientOpts))
		require.NoError(t, err, "failed to run experiment")
	})
}
