package client_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/mongo/readpref"
)

// Ensure that a non-existant mongoDB instance hits the server selection timeout
// rather than consuming resources indefinitely.
func TestServerSelectionTimeout(t *testing.T) {
	opts := options.Client().
		ApplyURI("mongodb://nonexistant:27017").
		SetServerSelectionTimeout(250 * time.Millisecond).
		SetConnectTimeout(1 * time.Second)

	client, err := mongo.Connect(opts)
	require.NoError(t, err)

	defer func() { assert.NoError(t, client.Disconnect(context.Background())) }()

	err = client.Ping(context.Background(), readpref.Primary())
	assert.Error(t, err)
}
