package client_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/mongo/readconcern"
)

func TestSnapshotReadConcernAppendedTwice(t *testing.T) {
	client, err := mongo.Connect()
	require.NoError(t, err)

	defer client.Disconnect(context.Background())

	sess, err := client.StartSession()
	require.NoError(t, err)

	defer sess.EndSession(context.Background())

	txnOpts := options.Transaction().SetReadConcern(readconcern.Snapshot())

	err = sess.StartTransaction(txnOpts)
	require.NoError(t, err)
}
