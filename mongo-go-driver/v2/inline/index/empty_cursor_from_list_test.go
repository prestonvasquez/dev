package index

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func TestEmptyCursorFromList(t *testing.T) {
	client, err := mongo.Connect(options.Client().ApplyURI(os.Getenv("MONGODB_URI")))
	require.NoError(t, err)

	defer func() {
		err := client.Disconnect(context.Background())
		require.NoError(t, err)
	}()

	coll := client.Database("db").Collection("vstore")
	view := coll.SearchIndexes()

	siOpts := options.SearchIndexes().SetName("x").SetType("vectorSearch")

	cursor, err := view.List(context.Background(), siOpts)
	require.NoError(t, err)

	fmt.Println(cursor.Current)
}
