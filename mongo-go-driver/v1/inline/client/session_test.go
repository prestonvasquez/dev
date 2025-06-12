package client_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readconcern"
)

func TestSnapshotReadConcernAppendedTwice(t *testing.T) {
	client, err := mongo.Connect(context.Background())
	require.NoError(t, err)

	defer client.Disconnect(context.Background())

	sess, err := client.StartSession()
	require.NoError(t, err)

	defer sess.EndSession(context.Background())

	txnOpts := options.Transaction().SetReadConcern(readconcern.Snapshot())

	err = sess.StartTransaction(txnOpts)
	require.NoError(t, err)

	err = mongo.WithSession(context.Background(), sess, func(sc mongo.SessionContext) error {
		collOpts := options.Collection().SetReadConcern(readconcern.Snapshot())
		coll := client.Database("testdb").Collection("test", collOpts)

		fmt.Println("1")
		// First command should succeed
		_, err = coll.Find(sc, bson.D{})
		require.NoError(t, err)

		// Second command should fial with a server error indicating that the read
		// concern has already been set
		_, err = coll.Find(sc, bson.D{})
		require.Error(t, err)

		return nil
	})

	require.NoError(t, err)
}
