package client_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
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

func TestMyRepositoryFunction(t *testing.T) {
	ctx := context.Background()

	//  Start up a MongoDB container
	req := testcontainers.ContainerRequest{
		Image:        "mongo:latest",
		ExposedPorts: []string{"27017/tcp"},
		WaitingFor:   wait.ForLog("Waiting for connections").WithStartupTimeout(30 * time.Second),
	}
	mongoC, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err)
	defer mongoC.Terminate(ctx)

	// Get mapped host & port
	host, err := mongoC.Host(ctx)
	require.NoError(t, err)
	port, err := mongoC.MappedPort(ctx, "27017")
	require.NoError(t, err)

	// Connect with the Go driver
	uri := fmt.Sprintf("mongodb://%s:%s", host, port.Port())
	client, err := mongo.Connect(options.Client().ApplyURI(uri))
	require.NoError(t, err)
	defer client.Disconnect(ctx)

	err = client.Ping(ctx, nil)
	require.NoError(t, err)

	//coll := client.Database("testdb").Collection("mycoll")

	//_, err = coll.InsertOne(ctx, bson.D{{"x", "y"}})
	//require.NoError(t, err)
}
