package failpoint

import (
	"context"
	"fmt"
	"log"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

// NewFailpointError will create a failpoint that errors a specific number of
// times. This function returns a funciton to close the failpoint.
func NewErrorN(ctx context.Context, client *mongo.Client, cmdName string, errCode int, n int) (func(), error) {
	// Create a failpoint that fails everytime we try to insert.
	admindb := client.Database("admin")
	doc := bson.D{
		{"configureFailPoint", "failCommand"},
		{"mode", bson.D{{"times", n}}},
		{"data",
			bson.D{
				{"errorCode", errCode},
				{"closeConnection", false},
				{"failCommands", bson.A{cmdName}},
			},
		},
	}

	err := admindb.RunCommand(ctx, doc).Err()
	if err != nil {
		return nil, fmt.Errorf("could not run fail point command: %w", err)
	}

	return func() {
		doc := bson.D{
			{"configureFailPoint", "failCommand"},
			{"mode", "off"},
		}
		err = admindb.RunCommand(ctx, doc).Err()
		if err != nil {
			log.Fatalf("could not disable fail point command: %v", err)
		}
	}, nil
}
