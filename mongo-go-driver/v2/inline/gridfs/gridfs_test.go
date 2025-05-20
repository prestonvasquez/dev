package gridfs_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func TestDecodingIntoID(t *testing.T) {
	client, err := mongo.Connect()
	require.NoError(t, err)

	defer client.Disconnect(context.Background())

	// Get the database and create a GridFS bucket
	db := client.Database("gridfs_test_db")

	// Drop the collection
	db.Collection("myfiles.files").Drop(context.Background())
	db.Collection("myfiles.chunks").Drop(context.Background())

	bucket := db.GridFSBucket(options.GridFSBucket().SetName("myfiles"))

	// Data to upload
	fileName := "example-file.txt"
	fileContent := []byte("Hello GridFS! This is a test file.")

	// Upload data into GridFS
	uploadStream, err := bucket.OpenUploadStream(context.Background(), fileName)
	require.NoError(t, err)

	_, err = uploadStream.Write(fileContent)
	require.NoError(t, err)

	uploadStream.Close()

	// Verify the file metadata
	fileCursor, err := bucket.Find(context.Background(), bson.D{})
	require.NoError(t, err)

	for fileCursor.Next(context.Background()) {
		var file mongo.GridFSFile
		err := fileCursor.Decode(&file)
		fmt.Println(file)
		require.NoError(t, err)

		assert.NotNil(t, file.ID)
	}
}
