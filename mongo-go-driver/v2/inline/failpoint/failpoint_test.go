package failpoint

import (
	"context"
	"fmt"
	"log"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

func TestInsertOneTimeout(t *testing.T) {
	client, err := mongo.Connect()
	require.NoError(t, err)

	defer func() { assert.NoError(t, client.Disconnect(context.Background())) }()

	td, err := createBlockFP(client, "insert", 500, 1)
	defer td()

	coll := client.Database("testdb").Collection("coll")
	defer func() { assert.NoError(t, coll.Drop(context.Background())) }()

	ctx, cancel := context.WithTimeout(context.Background(), 10)
	defer cancel()

	_, err = coll.InsertOne(ctx, bson.D{{"msg", "hello"}})
	fmt.Println("Error:", err)
	//assert.Error(t, err)
	//assert.ErrorIs(t, err, context.DeadlineExceeded)
}

func createBlockFP(client *mongo.Client, cmd string, blockTime, iter int) (func(), error) {
	admindb := client.Database("admin")

	// Create a document for the run command that sets a fail command that is always on.
	failCommand := bson.D{
		{Key: "configureFailPoint", Value: "failCommand"},
		{Key: "mode", Value: bson.D{{"times", iter}}},
		{Key: "data", Value: bson.D{
			{Key: "blockConnection", Value: true},
			{Key: "blockTimeMS", Value: blockTime},
			{Key: "failCommands", Value: bson.A{cmd}},
		}},
	}

	err := admindb.RunCommand(context.Background(), failCommand).Err()
	if err != nil {
		return func() {}, err
	}

	return func() {
		doc := bson.D{
			{Key: "configureFailPoint", Value: "failCommand"},
			{Key: "mode", Value: "off"},
		}

		err = admindb.RunCommand(context.Background(), doc).Err()
		if err != nil {
			log.Fatalf("could not disable fail point command: %v", err)
		}
	}, nil
}
