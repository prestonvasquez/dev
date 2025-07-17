package database_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func TestCreateCollection(t *testing.T) {
	const uri = "mongodb://127.0.0.1:28017/?directConnection=true"

	clientOpts := options.Client().ApplyURI(uri)
	client, err := mongo.Connect(clientOpts)
	//client, err := mongo.Connect()
	require.NoError(t, err)

	defer func() {
		err := client.Disconnect(context.Background())
		require.NoError(t, err)
	}()

	err = client.Database("test").CreateCollection(context.Background(), "x")
	require.NoError(t, err)
}

func TestRunCommandProxyTest(t *testing.T) {
	const uri = "mongodb://127.0.0.1:28017/?directConnection=true"

	clientOpts := options.Client().ApplyURI(uri)
	client, err := mongo.Connect(clientOpts)
	//client, err := mongo.Connect()
	require.NoError(t, err)

	defer func() {
		err := client.Disconnect(context.Background())
		require.NoError(t, err)
	}()

	proxyTest := bson.D{
		{Key: "actions", Value: bson.A{
			bson.D{{Key: "delayMs", Value: 400}},
			//bson.D{{Key: "sendBytes", Value: 0}},
			//bson.D{{Key: "sendAll", Value: true}},
		}},
	}

	type myStruct struct {
		Name string `bson:"name"`
		Age  int    `bson:"age"`
	}

	cmd := bson.D{
		//{Key: "insert", Value: "mycoll"},
		//{Key: "documents", Value: bson.A{myStruct{Name: "Alice", Age: 30}}},
		{Key: "ping", Value: 1},
		{Key: "proxyTest", Value: proxyTest},
	}

	db := client.Database("admin")

	// Run the command against the proxy with timeoutMS.
	err = db.RunCommand(context.Background(), cmd).Err()
	require.NoError(t, err)
}
