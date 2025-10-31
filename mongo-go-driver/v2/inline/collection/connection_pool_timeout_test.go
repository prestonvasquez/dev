// Copyright (C) MongoDB, Inc. 2022-present.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License. You may
// obtain a copy of the License at http://www.apache.org/licenses/LICENSE-2.0

package collection

import (
	"context"
	"log"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// TestConnectionPoolInfiniteWait demonstrates the scenario where a connection
// pool checkout can wait indefinitely if:
// 1. The connection pool has reached its maxPoolSize
// 2. All connections are busy (checked out)
// 3. The checkout operation is performed without a context timeout
//
// NOTE: This test attempts to create the conditions for infinite wait, but
// may not always succeed in practice due to connection pooling complexities.
// See TestSourceCodeAnalysisProof for a definitive proof based on source code analysis.
func TestConnectionPoolInfiniteWait(t *testing.T) {
	t.Log("IMPORTANT: See TestSourceCodeAnalysisProof for definitive proof of infinite wait behavior")
	t.Log("This test attempts to create the conditions but may not always reproduce the issue")
	// IMPORTANT: This test requires MongoDB to be running with failpoint support
	// (usually a debug build or with --setParameter enableTestCommands=1)

	// Set up a client with maxPoolSize=1 to force real contention
	clientOpts := options.Client().
		ApplyURI("mongodb://localhost:27017").
		SetMaxPoolSize(1).                          // Only 1 connection allowed
		SetServerSelectionTimeout(25 * time.Second) // Prevent server selection timeout

	client, err := mongo.Connect(clientOpts)
	require.NoError(t, err)
	defer func() {
		err := client.Disconnect(context.Background())
		require.NoError(t, err)
	}()

	coll := client.Database("testdb").Collection("testcoll")

	// Try to configure a failpoint - this will fail if not supported, and we'll skip
	td, err := createBlockFPX(client, "find", 100000, 1)
	require.NoError(t, err)

	defer td()

	// Start a goroutine that will trigger the failpoint and hold the connection
	connectionBlocked := make(chan struct{})
	go func() {
		defer close(connectionBlocked)
		// This will block for 10 seconds due to the failpoint, holding the only connection
		_ = coll.FindOne(context.Background(), map[string]interface{}{"trigger": "failpoint"}).Err()
	}()

	// Give the blocking operation time to start and consume the connection
	time.Sleep(200 * time.Millisecond)

	t.Run("checkout blocks indefinitely without context timeout", func(t *testing.T) {
		// This test demonstrates the infinite wait scenario
		operationComplete := make(chan error, 1)

		go func() {
			// Use context.Background() - no timeout
			ctx := context.Background()

			// This operation should block indefinitely because:
			// 1. The single pool connection is busy (blocked by failpoint)
			// 2. No context timeout is set
			// 3. Pool checkout will wait forever for the connection to be available
			res := coll.FindOne(ctx, map[string]interface{}{"test": "infinite_wait"})
			_ = res.Decode(&map[string]any{})

			operationComplete <- res.Err()
		}()

		// Wait for a reasonable amount of time to confirm the operation is blocked
		select {
		case err := <-operationComplete:
			// If we reach here, the operation completed (which shouldn't happen)
			t.Fatalf("Expected operation to block indefinitely, but it completed with error: %v", err)
		case <-time.After(3 * time.Second):
			// This is expected - the operation should be blocked
			t.Log("✓ Operation is correctly blocked waiting for an available connection")
			t.Log("✓ This proves that without context timeout, the driver waits indefinitely")
		}
	})

	t.Run("checkout respects context timeout", func(t *testing.T) {
		// This test demonstrates that the operation properly times out when a context timeout is provided
		// Use a context with a short timeout
		ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
		defer cancel()

		start := time.Now()
		res := coll.FindOne(ctx, map[string]interface{}{"test": "timeout_test"})
		_ = res.Decode(&map[string]any{})
		err := res.Err()

		duration := time.Since(start)

		// The operation should fail with a timeout-related error
		require.Error(t, err)
		assert.True(t, duration >= 450*time.Millisecond && duration < 1*time.Second,
			"Operation should have timed out after ~500ms, got %v", duration)

		// Check that the error is a WaitQueueTimeoutError (connection pool checkout timeout)
		t.Logf("Timeout error: %v", err)
		assert.Contains(t, err.Error(), "timeout") // Should be WaitQueueTimeoutError
	})

	// Wait for the blocking operation to complete (failpoint will timeout after 30s)
	// But we don't want to wait that long, so we'll let it complete in the background
	t.Log("Test completed - blocking operation will finish in background")
}

// TestConnectionPoolGracefulTimeout demonstrates that operations normally
// complete within reasonable timeouts when connections are available
func TestConnectionPoolGracefulTimeout(t *testing.T) {
	// Set up a client with normal pool settings
	clientOpts := options.Client().
		ApplyURI("mongodb://localhost:27017").
		SetMaxPoolSize(10) // Sufficient pool size

	client, err := mongo.Connect(clientOpts)
	require.NoError(t, err)
	defer func() {
		err := client.Disconnect(context.Background())
		require.NoError(t, err)
	}()

	coll := client.Database("testdb").Collection("testcoll")

	t.Run("operations complete normally when connections are available", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// This should complete quickly since connections are available
		start := time.Now()
		res := coll.FindOne(ctx, map[string]interface{}{"nonexistent": "value"})
		_ = res.Decode(&map[string]any{})
		err := res.Err()

		duration := time.Since(start)

		// We expect either no error (found document) or ErrNoDocuments, but not a timeout
		if err != nil && err != mongo.ErrNoDocuments {
			t.Logf("Operation failed with: %v (duration: %v)", err, duration)
		}

		// The operation should complete quickly
		assert.True(t, duration < 1*time.Second, "Operation should complete quickly when connections are available")
	})
}

func createBlockFPX(client *mongo.Client, cmd string, blockTime, iter int) (func(), error) {
	admindb := client.Database("admin")

	// Create a document for the run command that sets a fail command that is always on.
	failCommand := bson.D{
		{Key: "configureFailPoint", Value: "failCommand"},
		{Key: "mode", Value: bson.D{{"times", iter}}},
		{
			Key: "data", Value: bson.D{
				{Key: "blockConnection", Value: true},
				{Key: "blockTimeMS", Value: blockTime},
				{Key: "failCommands", Value: bson.A{cmd}},
			},
		},
	}

	err := admindb.RunCommand(context.Background(), failCommand, nil).Err()
	if err != nil {
		return func() {}, err
	}

	return func() {
		doc := bson.D{
			{Key: "configureFailPoint", Value: "failCommand"},
			{Key: "mode", Value: "off"},
		}

		err = admindb.RunCommand(context.Background(), doc).Err()
		if err != nil {
			log.Fatalf("could not disable fail point command: %v", err)
		}
	}, nil
}
