package client_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func TestPing(t *testing.T) {
	const uri = "mongodb://127.0.0.1:28017/?directConnection=true"

	clientOpts := options.Client().ApplyURI(uri)
	client, err := mongo.Connect(clientOpts)
	require.NoError(t, err)

	err = client.Ping(context.Background(), nil)
	require.NoError(t, err)
}

//func TestInternalClientOptions(t *testing.T) {
//	opts := options.Client()
//	_ = xoptions.SetInternalInsertOneOptions(opts, "rawData", true)
//}
