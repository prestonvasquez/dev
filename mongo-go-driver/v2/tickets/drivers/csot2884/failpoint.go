package csot2884

import (
	"context"
	"log"
	"testing"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

func createFPAlwaysOn(ctx context.Context, t *testing.T, client *mongo.Client, blockTime int) (func(), error) {
	admindb := client.Database("admin")

	// Create a document for the run command that sets a fail command that is always on.
	failCommand := bson.D{
		{Key: "configureFailPoint", Value: "failCommand"},
		{Key: "mode", Value: "alwaysOn"},
		{Key: "data", Value: bson.D{
			{Key: "blockConnection", Value: true},
			{Key: "blockTimeMS", Value: blockTime},
			{Key: "failCommands", Value: bson.A{"insert"}}},
		},
	}

	err := admindb.RunCommand(ctx, failCommand).Err()
	if err != nil {
		return func() {}, err
	}

	return func() {
		doc := bson.D{
			{Key: "configureFailPoint", Value: "failCommand"},
			{Key: "mode", Value: "off"},
		}

		err = admindb.RunCommand(ctx, doc).Err()
		if err != nil {
			log.Fatalf("could not disable fail point command: %v", err)
		}
	}, nil
}
