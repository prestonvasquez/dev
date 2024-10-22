package csot2884

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

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
