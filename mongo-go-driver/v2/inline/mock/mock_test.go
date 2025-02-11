package mypackage

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/x/mongo/driver/drivertest"
)

func TestMyRepositoryFunction(t *testing.T) {
	ctx := context.Background()
	//mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	//
	//
	mockDep := drivertest.NewMockDeployment()

	opts := options.Client()
	opts.Deployment = mockDep

	//opts.Deployment.AddResponses(bson.D{{Key: "ok", Value: 1}})

	client, err := mongo.Connect(opts)
	require.NoError(t, err)

	mockDep.AddResponses(
		bson.D{{Key: "ok", Value: "1"}},
		bson.D{
			{Key: "ok", Value: "1"},
			{Key: "value", Value: bson.M{"_id": "custom123", "key": 24}},
		})

	err := MyRepositoryFunction(ctx, client.Database().Collection("mycoll"))

	assert.NoError(t, err, "Should have successfully run")
}
