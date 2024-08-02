package indexutil

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// CreateN will create n arbitrary indexes.
func CreateN(ctx context.Context, coll *mongo.Collection, n int) error {
	models := make([]mongo.IndexModel, n)
	for i := 0; i < n; i++ {
		name := fmt.Sprintf("name_%v", i)

		models[i] = mongo.IndexModel{Keys: bson.D{{name, 1}}}
	}

	_, err := coll.Indexes().CreateMany(ctx, models)

	return err
}
