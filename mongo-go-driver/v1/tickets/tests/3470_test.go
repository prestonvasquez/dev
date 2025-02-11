package tests

import (
	"errors"
	"fmt"
	"testing"

	"github.com/RoaringBitmap/roaring"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/bsontype"
	"go.mongodb.org/mongo-driver/x/bsonx/bsoncore"
)

type MockStruct struct {
	ID        string        `bson:"_id"`
	Overrider MockOverrider `bson:"overrider"`
}

type MockOverrider struct {
	bitmap *roaring.Bitmap
}

func (m MockOverrider) MarshalBSONValue() (bsontype.Type, []byte, error) {
	if m.bitmap == nil || m.bitmap.IsEmpty() {
		return bson.TypeNull, nil, nil
	}
	bytes, err := m.bitmap.MarshalBinary()
	if err != nil {
		return bson.TypeUndefined, nil, err
	}
	return bson.MarshalValue(bytes)
}

func (m *MockOverrider) UnmarshalBSONValue(t bsontype.Type, raw []byte) error {
	fmt.Println(m == nil)
	m.bitmap = roaring.New()
	if t == bson.TypeNull {
		return nil
	}
	if t != bson.TypeBinary {
		return fmt.Errorf("unable to unmarshal Bitmap from bson type: %v", t)
	}

	_, bytes, _, ok := bsoncore.ReadBinary(raw)
	if !ok {
		return errors.New("unable to ReadBinary to unmarshal Bitmap")
	}

	err := m.bitmap.UnmarshalBinary(bytes)
	if err != nil {
		return fmt.Errorf("unable to unmarshal bson byte array to unmarshal Bitmap: %w", err)
	}

	return nil
}

func TestUnmarshalNilBehavior(t *testing.T) {
	// marshal a mock struct with a nil bitmap
	expected := MockStruct{
		ID: "test",
	}
	expectedBytes, err := bson.Marshal(&expected)
	require.Nil(t, err)

	// unmarshal
	var found MockStruct
	err = bson.Unmarshal(expectedBytes, &found)
	require.Nil(t, err)

	// bitmap should not be nil
	assert.NotNil(t, found.Overrider.bitmap)
}
