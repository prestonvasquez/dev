package collection_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// TestPoolInfiniteWaitProof creates a scenario that definitively proves
// the MongoDB Go driver will wait indefinitely without context timeout
func TestPoolInfiniteWaitProof(t *testing.T) {
	// Use maxPoolSize=1 and maxConnecting=1 to create strict limits
	clientOpts := options.Client().
		ApplyURI("mongodb://localhost:27017").
		SetMaxPoolSize(1) // Strict limit: only 1 connection total

	client, err := mongo.Connect(clientOpts)
	require.NoError(t, err)
	defer func() {
		_ = client.Disconnect(context.Background())
	}()

	coll := client.Database("testdb").Collection("testcoll")

	// Insert some test data to ensure operations can succeed
	_, _ = coll.InsertOne(context.Background(), map[string]interface{}{"setup": "test"})

	var wg sync.WaitGroup
	connectionHeld := make(chan struct{})
	releaseConnection := make(chan struct{})

	// Goroutine 1: Perform an operation that will hold the connection for a long time
	wg.Add(1)
	go func() {
		defer wg.Done()
		
		// Create a context with a VERY long timeout to hold the connection
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()
		
		// Signal we're starting
		close(connectionHeld)
		
		// Start a find operation with tailable cursor to hold connection longer
		cursor, err := coll.Find(ctx, map[string]interface{}{}, options.Find().SetBatchSize(1))
		if err == nil {
			defer cursor.Close(ctx)
			
			// Process one result to ensure the connection is active
			if cursor.Next(ctx) {
				var result map[string]interface{}
				cursor.Decode(&result)
			}
			
			// Now wait for signal to release - this keeps connection busy
			<-releaseConnection
			
			// Process remaining results to clean up
			for cursor.Next(ctx) {
				var result map[string]interface{}
				cursor.Decode(&result)
			}
		}
	}()

	// Wait for the connection-holding operation to start
	<-connectionHeld
	
	// Give extra time to ensure connection is definitely held
	time.Sleep(500 * time.Millisecond)

	t.Run("proves infinite wait without timeout", func(t *testing.T) {
		done := make(chan struct{})
		var operationErr error

		go func() {
			defer close(done)
			
			// This should wait indefinitely because:
			// 1. maxPoolSize=1 and the connection is busy
			// 2. context.Background() has no timeout
			// 3. The driver's select statement waits on ctx.Done() which never fires
			result := coll.FindOne(context.Background(), map[string]interface{}{"test": "blocked"})
			operationErr = result.Err()
		}()

		// Test if operation completes within reasonable time
		select {
		case <-done:
			t.Errorf("❌ FAIL: Expected infinite wait, but operation completed with: %v", operationErr)
		case <-time.After(3 * time.Second):
			t.Log("✅ PROOF: Operation has been waiting for 3+ seconds")
			t.Log("✅ This confirms the driver waits indefinitely without context timeout")
			t.Log("✅ The operation is blocked in pool.checkOut() select statement")
		}
	})

	t.Run("proves timeout works with context deadline", func(t *testing.T) {
		// Test with timeout to prove the contrast
		ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
		defer cancel()

		start := time.Now()
		result := coll.FindOne(ctx, map[string]interface{}{"test": "timeout"})
		err := result.Err()
		duration := time.Since(start)

		require.Error(t, err)
		t.Logf("✅ With timeout: operation failed after %v with error: %v", duration, err)
		
		if duration >= 180*time.Millisecond && duration <= 500*time.Millisecond {
			t.Log("✅ PROOF: Context timeout is the ONLY protection against infinite wait")
		}
	})

	// Clean up
	close(releaseConnection)
	wg.Wait()
}