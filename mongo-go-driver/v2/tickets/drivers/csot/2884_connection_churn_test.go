package csot

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

func Test2884_SimplePath(t *testing.T) {
	// Create a client that captures and logs the following to IO:
	//
	//   (1) Connection check out
	//   (2) Connection check in
	//   (3) Connection closed
	//   (4) Command started
	//   (5) Command succeeded
	//   (6) Command failed
	//   (7) Pending read started
	//   (8) Pending read succeeded
	//   (9) Pending read failed
	sink := newCustomLogger(t, true)

	loggerOptions := options.
		Logger().
		SetSink(sink).
		SetComponentLevel(options.LogComponentCommand, options.LogLevelDebug).
		SetComponentLevel(options.LogComponentConnection, options.LogLevelDebug)

	opts := options.Client().SetLoggerOptions(loggerOptions).SetMaxPoolSize(1)

	client, err := mongo.Connect(opts)
	require.NoError(t, err, "failed to connect to server")

	defer func() {
		_ = client.Disconnect(context.Background())
	}()

	// Create a failpoint to block for theh first operation and the first
	// pending read.
	teardown, err := createBlockFP(t, client, "insert", 250, 1)
	require.NoError(t, err, "failed to create blocking failpoint")

	defer teardown()

	coll := client.Database("db").Collection("coll")

	sink.toggleOn() // turn logging on

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	_, err = coll.InsertMany(ctx, []bson.D{bson.D{{"y", 1}}})
	assert.ErrorIs(t, err, context.DeadlineExceeded)

	ctx, cancel = context.WithTimeout(context.Background(), 150*time.Millisecond)
	defer cancel()

	_, err = coll.InsertMany(ctx, []bson.D{bson.D{{"y", 1}}})
	assert.Error(t, err) // You will get a timeout since the default allowable time for pending read is 400 < ~700

	_, err = coll.InsertMany(context.Background(), []bson.D{bson.D{{"y", 1}}})
	assert.NoError(t, err)

	sink.toggleOff() // turn logging off

	assert.Equal(t, sink.commandStartedCount, 2)                 // Once for first, once for last
	assert.Equal(t, sink.commandFailedCount, 1)                  // Once for first
	assert.Equal(t, sink.commandSucceededCount, 1)               // Once for last
	assert.Equal(t, sink.connectionPendingReadStartedCount, 2)   // Once for second, once for last
	assert.Equal(t, sink.connectionPendingReadFailedCount, 1)    // Once for second
	assert.Equal(t, sink.connectionPendingReadSucceededCount, 1) // Once for last
	assert.Equal(t, sink.connectionCheckedOutCount, 2)           // Once for first, once for last (silent if pending fails)
	assert.Equal(t, sink.connectionCheckedInCount, 2)            // Once for first, once for last (silent if pending fails)
	assert.Equal(t, sink.connectionClosedCount, 0)               // No connections should be closed
}
