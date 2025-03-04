package collection

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func TestRenameCollection(t *testing.T) {
	client, err := mongo.Connect(context.Background())
	require.NoError(t, err)

	defer func() {
		err := client.Disconnect(context.Background())
		require.NoError(t, err)
	}()

	db := client.Database("testdb")

	coll := db.Collection("coll")
	coll1 := db.Collection("coll1")

	coll.Drop(context.Background())
	coll1.Drop(context.Background())

	err = db.CreateCollection(context.Background(), "coll")
	require.NoError(t, err)

	adminDB := client.Database("admin")

	res := adminDB.RunCommand(context.Background(), bson.D{
		{"renameCollection", "testdb.coll"},
		{"to", "testdb.coll1"},
		{"dropTarget", true},
		{"comment", "Report platform"},
	})

	err = res.Err()
	require.NoError(t, err)
}
