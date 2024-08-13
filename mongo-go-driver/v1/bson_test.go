package v1

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func Test_NumericUnmarshaling(t *testing.T) {
	// GODRIVER-3296
	t.Run("int32 to decimal128 bson", func(t *testing.T) {
		doc := bson.D{{"test", int32(1)}}

		bytes, err := bson.Marshal(doc)
		assert.NoError(t, err)

		result := struct {
			Test primitive.Decimal128 `bson:"test"`
		}{}

		err = bson.Unmarshal(bytes, &result)
		assert.Error(t, err) // Cannot decode int32 to decimal128
	})

	t.Run("int32 to float64 bson", func(t *testing.T) {

		doc := bson.D{{"test", int32(1)}}
		bytes, err := bson.Marshal(doc)
		assert.NoError(t, err)

		result := struct {
			Test float64 `bson:"test"`
		}{}

		err = bson.Unmarshal(bytes, &result)

		assert.NoError(t, err)
		assert.Equal(t, 1.0, result.Test)
	})
}
