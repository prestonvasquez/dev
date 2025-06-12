package csot

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func loadLargeCollectionForClient(t *testing.T, coll *mongo.Collection) {
	const volume = 1000
	const goRoutines = 10
	const batchSize = 500 // Size of batches to load for testing

	docs := make([]interface{}, 0, batchSize)
	for i := 0; i < batchSize; i++ {
		docs = append(docs, bson.D{
			{"field1", rand.Int63()},
			{"field2", rand.Int31()},
		})
	}

	perGoroutine := volume / goRoutines

	errs := make(chan error, goRoutines*perGoroutine)
	done := make(chan struct{}, goRoutines)

	for i := 0; i < int(goRoutines); i++ {
		go func(i int) {
			for j := 0; j < perGoroutine; j++ {
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

		for i := 0; i < int(goRoutines); i++ {
			<-done
		}
	}()

	// Await errors and return the first error encountered.
	for err := range errs {
		require.NoError(t, err)
	}
}

func Test2884_ClientDisconnectBlocks(t *testing.T) {
	monitor := newMonitor(true, "find")

	opts := options.Client().
		SetMaxPoolSize(1).
		SetMonitor(monitor.commandMonitor)

	client, err := mongo.Connect(opts)
	require.NoError(t, err, "failed to connect to server")

	// 1. Insert a huge amount of data

	unindexedColl := client.Database("testdb").Collection("coll")
	//defer func() { unindexedColl.Drop(context.Background()) }()

	// Load the test data.
	loadLargeCollectionForClient(t, unindexedColl)

	// 2. Set low t/o on the operation to query it
	query := bson.D{{"field1", "doesntexist"}}

	const threshold = 100

	t.Log("starting op iteration")
	for i := 0; i < threshold; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
		defer cancel()

		err = unindexedColl.FindOne(ctx, query).Err()
		assert.ErrorIs(t, err, context.DeadlineExceeded)

		// 3. disconnect t/o and see what happens
		t.Logf("about to disconnect: %v\n", i)

		err = client.Disconnect(context.Background())
		require.NoError(t, err)
	}
}

func Test2884_CloseWhenNoRemainingTime(t *testing.T) {
	monitor := newMonitor(true, "insert")

	opts := options.Client().
		SetMonitor(monitor.commandMonitor).
		SetPoolMonitor(monitor.poolMonitor).
		SetMaxPoolSize(1)

	client, err := mongo.Connect(opts)
	require.NoError(t, err, "failed to connect to server")

	defer func() {
		t.Log("about to disconnect client")
		_ = client.Disconnect(context.Background())
	}()

	// Create a failpoint that will block 1 time for 150ms.
	teardown, err := createBlockFP(t, client, "insert", 750, 1)
	require.NoError(t, err, "failed to create blocking failpoint")

	defer teardown()

	coll := client.Database("db").Collection("coll")

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	st := time.Now()
	_, err = coll.InsertMany(ctx, []bson.D{bson.D{{"y", 1}}})
	fmt.Println("elapsed: ", time.Since(st))
	assert.ErrorIs(t, err, context.DeadlineExceeded)

	ctx, cancel = context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	_, err = coll.InsertMany(ctx, []bson.D{bson.D{{"y", 1}}})
	assert.ErrorIs(t, err, context.DeadlineExceeded)

	// Expected events
	assert.Len(t, monitor.commandStarted, 1)
	assert.Len(t, monitor.commandFailed, 1)
	//assert.Len(t, monitor.connectionPendingReadStarted, 1)
	assert.Len(t, monitor.connectionClosed, 1)
}

func Test2884_ConcurrentOps(t *testing.T) {
	monitor := newMonitor(false, "insert")

	opts := options.Client().
		SetMonitor(monitor.commandMonitor).
		SetPoolMonitor(monitor.poolMonitor).
		SetMaxPoolSize(1)

	client, err := mongo.Connect(opts)
	require.NoError(t, err, "failed to connect to server")

	defer func() {
		_ = client.Disconnect(context.Background())
	}()

	// Create a failpoint that will block 1 time for 150ms.
	teardown, err := createBlockFP(t, client, "insert", 200, 1)
	require.NoError(t, err, "failed to create blocking failpoint")

	defer teardown()

	coll := client.Database("db").Collection("coll")

	monitor.Reset()

	threshold := 200

	wg := sync.WaitGroup{}
	wg.Add(threshold)

	for i := 0; i < threshold; i++ {
		go func() {
			defer wg.Done()

			ctx, cancel := context.WithTimeout(context.Background(), 75*time.Millisecond)
			defer cancel()

			_, _ = coll.InsertOne(ctx, bson.D{})
			//assert.ErrorIs(t, err, context.DeadlineExceeded, "iter %d was not a CSOT error", i)
		}()
	}

	// Expect 3 check outs and 3 check ins
	wg.Wait()
}

// Ensure that re-use of a connection for a pending read doesn't depend on
// checking that connection back into the pool.
func Test2884_CheckInState(t *testing.T) {
	monitor := newMonitor(true, "insert")

	opts := options.Client().
		SetMonitor(monitor.commandMonitor).
		SetPoolMonitor(monitor.poolMonitor).
		SetMaxPoolSize(1)

	client, err := mongo.Connect(opts)
	require.NoError(t, err, "failed to connect to server")

	defer func() {
		_ = client.Disconnect(context.Background())
	}()

	// Create a failpoint that will block 1 time for 450ms.
	teardown, err := createBlockFP(t, client, "insert", 450, 1)
	require.NoError(t, err, "failed to create blocking failpoint")

	defer teardown()

	coll := client.Database("db").Collection("coll")

	monitor.Reset()

	// Insert a document three times, each with a timeout of 200ms.
	//
	//   - First time should time out sending the message
	//   - Second time should time out attempting the pending read in check out
	//   - Third time succeeds
	//
	// We expect the same connection to be used for all operations.
	for i := 1; i < 4; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)

		_, err := coll.InsertOne(ctx, bson.D{})
		if i < 3 {
			assert.ErrorIs(t, err, context.DeadlineExceeded, "iter %d was not a CSOT error", i)
		} else {
			require.NoError(t, err, "insert for iter %d failed", i)
		}

		cancel()
	}

	// Expect only one connection to ever be checked out/in
	assert.Len(t, monitor.connectionCheckedOut, 1)
	assert.Len(t, monitor.connectionCheckedIn, 1)

	var connID int64
	for id := range monitor.connectionCheckedOut {
		connID = id
		break
	}

	// Expect 2 check outs and 2 check ins
	assert.Len(t, monitor.connectionCheckedOut[connID], 2)
	assert.Len(t, monitor.connectionCheckedIn[connID], 2)

	//// Expect 2 await pending reads
	//assert.Len(t, monitor.connectionPendingReadStarted[connID], 2)

	//// Expect 1 pending read to fail
	//assert.Len(t, monitor.connectionPendingReadFailed[connID], 1)

	//// Expect 1 pending read to succeed
	//assert.Len(t, monitor.connectionPendingReadSucceeded[connID], 1)
}

func Test_3006_ChangeStream(t *testing.T) {
	opts := options.Client()

	client, err := mongo.Connect(opts)
	require.NoError(t, err, "failed to connect to server")

	defer func() {
		t.Log("about to disconnect")
		_ = client.Disconnect(context.Background())
	}()

	coll := client.Database("db").Collection("coll")

	csOpts := options.ChangeStream().SetBatchSize(1)

	cs, err := coll.Watch(context.Background(), mongo.Pipeline{}, csOpts)
	require.NoError(t, err)

	defer cs.Close(context.Background())

	ctx, cancel := context.WithTimeout(context.Background(), 150*time.Millisecond)
	defer cancel()

	ok := cs.Next(ctx)
	assert.False(t, ok)
	assert.ErrorIs(t, cs.Err(), context.DeadlineExceeded)

	// Insert some documents
	for i := 0; i < 2; i++ {
		_, err = coll.InsertOne(context.Background(), bson.D{})
		require.NoError(t, err)
	}

	// A subsequent call to next that does not time out should resume since the
	// error was CSOT.
	ok = cs.Next(context.Background())
	assert.True(t, ok)
}

// What happens to the connection when maxTimeMS expires?
func TestMaxTimeMSBehavior(t *testing.T) {
	monitor := newMonitor(true, "insert")

	opts := options.Client().
		SetPoolMonitor(monitor.poolMonitor).
		SetMonitor(monitor.commandMonitor).
		SetMaxPoolSize(1)

	client, err := mongo.Connect(opts)
	require.NoError(t, err, "failed to connect to server")

	defer func() {
		t.Log("about to disconnect")
		_ = client.Disconnect(context.Background())
	}()

	coll := client.Database("db").Collection("coll")

	err = coll.Drop(context.Background())
	require.NoError(t, err, "failed to drop test collection")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Millisecond)
	defer cancel()

	coll.InsertOne(ctx, bson.D{})
}

// Foreground: If an operation takes longer than the pending read timeout, then
// ensure there are a finite number of calls to "check out" that would result in
// successfully reading the pending data.
//
// TODO: Use driverutil package to mock the connection.
func TestOpTimeGTPendingReadLimit(t *testing.T) {
	const opTimeMS = 500

	monitor := newMonitor(false, "insert")

	opts := options.Client().
		SetPoolMonitor(monitor.poolMonitor).
		SetMonitor(monitor.commandMonitor).
		SetMaxPoolSize(1)

	client, err := mongo.Connect(opts)
	require.NoError(t, err, "failed to connect to server")

	defer func() {
		_ = client.Disconnect(context.Background())
	}()

	teardown, err := createFPAlwaysOn(context.Background(), t, client, opTimeMS)
	require.NoError(t, err, "failed to configure blocking failpoint")

	defer teardown()

	monitor.Reset()

	coll := client.Database("db").Collection("coll")

	err = coll.Drop(context.Background())
	require.NoError(t, err, "failed to drop test collection")

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	for {
		_, err = coll.InsertOne(ctx, bson.D{})
		assert.ErrorIs(t, err, context.DeadlineExceeded)
	}

	//time.Sleep((opTimeMS + 25) * time.Millisecond) // Wait to see if conn closes in bg/fg

	//// Expect no closures
	//assert.Len(t, monitor.connectionClosed, 1)
}

// Ensure that the connection does not close if the bg/fg time is greater than
// the time it takes to complete the operation.
func TestOpTimeLTPendingReadLimit(t *testing.T) {
	const opTimeMS = 75

	monitor := newMonitor(false, "insert")

	opts := options.Client().
		SetPoolMonitor(monitor.poolMonitor).
		SetMonitor(monitor.commandMonitor).
		SetMaxPoolSize(1)

	client, err := mongo.Connect(opts)
	require.NoError(t, err, "failed to connect to server")

	defer func() {
		_ = client.Disconnect(context.Background())
	}()

	teardown, err := createFPAlwaysOn(context.Background(), t, client, opTimeMS)
	require.NoError(t, err, "failed to configure blocking failpoint")

	defer teardown()

	monitor.Reset()

	coll := client.Database("db").Collection("coll")

	err = coll.Drop(context.Background())
	require.NoError(t, err, "failed to drop test collection")

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	_, err = coll.InsertOne(ctx, bson.D{})
	assert.ErrorIs(t, err, context.DeadlineExceeded)

	time.Sleep((opTimeMS + 25) * time.Millisecond) // Wait to see if conn closes in bg/fg

	// Expect no closures
	assert.Zero(t, len(monitor.connectionClosed))
}

func Test2884_NoDeadline(t *testing.T) {
	client, err := mongo.Connect()
	require.NoError(t, err, "failed to connect to server")

	defer func() {
		_ = client.Disconnect(context.Background())
	}()

	// Create a failpoint that will block 1 time for 150ms.
	teardown, err := createBlockFP(t, client, "insert", 60_000, 1)
	require.NoError(t, err, "failed to create blocking failpoint")

	defer teardown()

	coll := client.Database("db").Collection("coll")

	st := time.Now()
	_, err = coll.InsertMany(context.Background(), []bson.D{bson.D{{"y", 1}}})
	fmt.Println("elapsed: ", time.Since(st))
}
