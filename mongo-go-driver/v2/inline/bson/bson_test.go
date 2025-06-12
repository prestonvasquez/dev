package bson_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

func TestMarshalNil(t *testing.T) {
	var v any = nil
	_, err := bson.Marshal(v)
	require.Error(t, err) // Bug? GODRIVER-2786
}

func TestMarshalValueNil(t *testing.T) {
	var v any = nil
	_, _, err := bson.MarshalValue(v)
	fmt.Println(err)
	require.Error(t, err)
}

func TestDoubleDecoding(t *testing.T) {
	client, err := mongo.Connect()
	require.NoError(t, err)

	defer func() {
		err := client.Disconnect(context.Background())
		require.NoError(t, err)
	}()

	coll := client.Database("test").Collection("test")

	type Data struct {
		Content any `bson:"content"`
	}

	doc := Data{
		Content: []interface{}{1, 2, 3},
	}

	_, err = coll.InsertOne(context.Background(), doc)
	if err != nil {
		t.Fatalf("InsertOne failed: %v", err)
	}

	var result Data
	err = coll.FindOne(context.Background(), bson.D{}).Decode(&result)
	if err != nil {
		t.Fatalf("FindOne failed: %v", err)
	}

	result = Data{
		Content: []interface{}{1, 2, 8},
	}

	err = coll.FindOne(context.Background(), bson.D{}).Decode(&result)
	if err != nil {
		t.Fatalf("FindOne number TWO failed: %v", err)
	}

	fmt.Println(result)
}
func TestDoubleDecoding(t *testing.T) {
	client, err := mongo.Connect(...)
	require.NoError(t, err)

	defer func() {
		err := client.Disconnect(context.Background())
		require.NoError(t, err)
	}()

	coll := client.Database("test").Collection("test")

	type Data struct {
		Content []interface{} `bson:"content"`
	}

	doc := Data{
		Content: []interface{}{1, 2, 3},
	}

	_, err = coll.InsertOne(context.Background(), doc)
	if err != nil {
		t.Fatalf("InsertOne failed: %v", err)
	}

	t.Run("case1", func(t *testing.T) {
		result := Data{
			Content: []interface{}{7, 8, 9},
		}
		err = coll.FindOne(context.Background(), bson.D{}).Decode(&result)
		if err != nil {
			t.Fatalf("FindOne failed: %v", err)
		}

		t.Log(result)
	})
	t.Run("case2", func(t *testing.T) {
		v := 7
		result := Data{
			Content: []interface{}{&v},
		}
		err = coll.FindOne(context.Background(), bson.D{}).Decode(&result)
		if err != nil {
			t.Fatalf("FindOne failed: %v", err)
		}

		t.Log(result)
	})
}
