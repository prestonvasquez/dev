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

// TestInfiniteWaitDemonstration is a simple demonstration that the MongoDB Go driver
// WILL wait indefinitely for connection pool checkout if no context timeout is set
// and all connections are busy.
func TestInfiniteWaitDemonstration(t *testing.T) {
	// Set up a client with maxPoolSize=1 to force contention
	clientOpts := options.Client().
		ApplyURI("mongodb://localhost:27017").
		SetMaxPoolSize(1) // Only 1 connection allowed

	client, err := mongo.Connect(clientOpts)
	require.NoError(t, err)
	defer func() {
		_ = client.Disconnect(context.Background())
	}()

	coll := client.Database("testdb").Collection("testcoll")

	// Strategy: Use a controllable blocking operation to hold the connection
	var wg sync.WaitGroup
	holdConnection := make(chan struct{})
	releaseConnection := make(chan struct{})

	// Goroutine 1: Hold the only connection indefinitely
	wg.Add(1)
	go func() {
		defer wg.Done()
		
		// Signal that we're about to start
		close(holdConnection)
		
		// This creates a cursor that we keep open to hold the connection
		// We use a very large timeout to ensure it doesn't fail quickly
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		
		cursor, err := coll.Find(ctx, map[string]interface{}{})
		if err == nil {
			defer cursor.Close(ctx)
			
			// Wait for signal to release the connection
			// Meanwhile, the cursor is open and the connection is busy
			<-releaseConnection
			
			// Drain the cursor to complete the operation
			for cursor.Next(ctx) {
				// Process results if any
			}
		}
	}()

	// Wait for the connection-holding operation to start
	<-holdConnection
	
	// Give it time to actually acquire the connection
	time.Sleep(100 * time.Millisecond)

	t.Run("demonstrates infinite wait without context timeout", func(t *testing.T) {
		operationComplete := make(chan error, 1)

		go func() {
			// This operation will wait indefinitely because:
			// 1. The pool has maxPoolSize=1 and that connection is busy
			// 2. context.Background() has no timeout
			// 3. The driver will wait forever for the connection to become available
			res := coll.FindOne(context.Background(), map[string]interface{}{"test": "infinite"})
			operationComplete <- res.Err()
		}()

		// Check if the operation completes within a reasonable time
		select {
		case err := <-operationComplete:
			t.Errorf("Expected operation to wait indefinitely, but it completed with: %v", err)
		case <-time.After(2 * time.Second):
			t.Log("✓ SUCCESS: Operation is waiting indefinitely (as expected)")
			t.Log("✓ This proves the driver has NO built-in timeout for connection checkout")
		}
	})

	t.Run("demonstrates timeout when context has deadline", func(t *testing.T) {
		// Use a context with a short timeout
		ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
		defer cancel()

		start := time.Now()
		res := coll.FindOne(ctx, map[string]interface{}{"test": "timeout"})
		err := res.Err()
		duration := time.Since(start)

		// Should get a timeout error
		require.Error(t, err)
		if duration >= 400*time.Millisecond && duration <= 1*time.Second {
			t.Logf("✓ SUCCESS: Operation timed out after %v (as expected)", duration)
			t.Logf("✓ Error was: %v", err)
		} else {
			t.Errorf("Expected timeout around 500ms, got %v", duration)
		}
	})

	// Clean up: release the connection
	close(releaseConnection)
	wg.Wait()
}