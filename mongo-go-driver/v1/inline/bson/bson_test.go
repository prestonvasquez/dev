package bson_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
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
