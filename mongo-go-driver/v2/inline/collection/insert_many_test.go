package collection_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

func TestReflectionFreeInsertMany(t *testing.T) {
	docs := []bson.D{
		{{"x", 1}},
		{{"y", 2}},
	}

	rawDocs, err := rfConversion(docs)
	require.NoError(t, err, "rfConversion failed")

	fmt.Println("Raw documents:", rawDocs)

	client, err := mongo.Connect()
	require.NoError(t, err, "Failed to connect to MongoDB")

	defer func() {
		err := client.Disconnect(context.Background())
		require.NoError(t, err, "Failed to disconnect from MongoDB")
	}()

	coll := client.Database("testdb").Collection("testcoll")

	_, err = coll.InsertMany(context.Background(), rawDocs)
	require.NoError(t, err, "InsertMany failed")
}

// rfConversion converts a slice of documents into a slice of bson.Raw
// documents.
func rfConversion(documents interface{}) ([]any, error) {
	bytes, err := bson.Marshal(struct {
		Arr any `bson:"arr"`
	}{Arr: documents})
	if err != nil {
		panic(err)
	}

	raw := bson.Raw(bytes).Lookup("arr")

	var docSlice []any
	switch raw.Type {
	case bson.TypeArray:
		elems, err := bson.Raw(raw.Array()).Elements()
		if err != nil {
			panic(err)
		}

		docSlice = make([]any, len(elems))
		for idx, elem := range elems {
			docSlice[idx] = elem.Value().Document()
		}
	}

	return docSlice, nil
}
