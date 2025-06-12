package bson_test

import (
	"log"
	"testing"

	bson1 "go.mongodb.org/mongo-driver/bson"
	bson2 "go.mongodb.org/mongo-driver/v2/bson"
)

//func TestUnmarshal_OverwritesPrepopulatedSlice(t *testing.T) {
//	type Data struct {
//		Content []interface{} `bson:"content"`
//	}
//
//	doc := Data{
//		Content: []interface{}{1, 2, 3},
//	}
//
//	receiver := Data{
//		Content: []interface{}{7, 8},
//	}
//
//	t.Run("go.mongodb.org/mongo-driver/bson v1", func(t *testing.T) {
//		err := bson1.Unmarshal(docToBsonBytes(doc), &receiver)
//		require.NoError(t, err)
//
//		assert.Equal(t, []interface{}{int32(1), int32(2), int32(3)}, receiver.Content,
//			"Unmarshalled content should overwrite prepopulated slice")
//	})
//
//	t.Run("go.mongodb.org/mongo-driver/v2/bson v2", func(t *testing.T) {
//		err := bson2.Unmarshal(docToBsonBytes(doc), &receiver)
//		require.NoError(t, err)
//
//		assert.Equal(t, []interface{}{int32(1), int32(2), int32(3)}, receiver.Content,
//			"Unmarshalled content should overwrite prepopulated slice")
//	})
//
//	t.Run("encoding/json", func(t *testing.T) {
//		err := json.Unmarshal(docToJsonBytes(doc), &receiver)
//		require.NoError(t, err)
//
//		assert.Equal(t, []interface{}{float64(1), float64(2), float64(3)}, receiver.Content,
//			"Unmarshalled content should overwrite prepopulated slice")
//	})
//}
//
//func TestUnmarshal_OverwritesPrepopulatedSettableSlice(t *testing.T) {
//	type Data struct {
//		Content []interface{} `bson:"content"`
//	}
//
//	doc := Data{
//		Content: []interface{}{1},
//	}
//
//	seven := 7
//
//	receiver := Data{
//		Content: []interface{}{&seven},
//	}
//
//	t.Run("go.mongodb.org/mongo-driver/bson v1", func(t *testing.T) {
//		err := bson1.Unmarshal(docToBsonBytes(doc), &receiver)
//		require.NoError(t, err)
//
//		assert.Equal(t, []interface{}{int32(1)}, receiver.Content,
//			"Unmarshalled content should overwrite prepopulated slice")
//	})
//
//	t.Run("go.mongodb.org/mongo-driver/v2/bson v2", func(t *testing.T) {
//		err := bson2.Unmarshal(docToBsonBytes(doc), &receiver)
//		require.NoError(t, err)
//
//		assert.Equal(t, []interface{}{int32(1)}, receiver.Content,
//			"Unmarshalled content should overwrite prepopulated slice")
//	})
//
//	t.Run("encoding/json", func(t *testing.T) {
//		err := json.Unmarshal(docToJsonBytes(doc), &receiver)
//		require.NoError(t, err)
//
//		assert.Equal(t, []interface{}{float64(1)}, receiver.Content,
//			"Unmarshalled content should overwrite prepopulated slice")
//	})
//}
//
//func TestUnmarshal_OverwritesPrepopulatedMap(t *testing.T) {
//	type Data struct {
//		Content map[string]interface{} `bson:"content"`
//	}
//
//	doc := Data{
//		Content: map[string]interface{}{
//			"x": 1,
//			"y": 2,
//		},
//	}
//
//	receiver := Data{
//		Content: map[string]interface{}{
//			"x": 99,
//			"y": 42,
//		},
//	}
//
//	t.Run("go.mongodb.org/mongo-driver/bson v1", func(t *testing.T) {
//		err := bson1.Unmarshal(docToBsonBytes(doc), &receiver)
//		require.NoError(t, err)
//
//		assert.Equal(t,
//			map[string]interface{}{"x": int32(1), "y": int32(2)},
//			receiver.Content,
//			"Unmarshalled content should overwrite prepopulated map")
//	})
//
//	t.Run("go.mongodb.org/mongo-driver/v2/bson v2", func(t *testing.T) {
//		err := bson2.Unmarshal(docToBsonBytes(doc), &receiver)
//		require.NoError(t, err)
//
//		assert.Equal(t,
//			map[string]interface{}{"x": int32(1), "y": int32(2)},
//			receiver.Content,
//			"Unmarshalled content should overwrite prepopulated map")
//	})
//
//	t.Run("encoding/json", func(t *testing.T) {
//		err := json.Unmarshal(docToJsonBytes(doc), &receiver)
//		require.NoError(t, err)
//
//		// JSON numbers become float64
//		assert.Equal(t,
//			map[string]interface{}{"x": float64(1), "y": float64(2)},
//			receiver.Content,
//			"Unmarshalled content should overwrite prepopulated map")
//	})
//}
//
//func TestOverlongNull(t *testing.T) {
//	s := string([]byte{0xC0, 0x80}) // overlong NUL in invalid UTF-8
//
//	v := map[string]interface{}{
//		"data": s,
//	}
//	data, err := bson2.Marshal(v)
//	fmt.Printf("err: %v\n", err)
//	fmt.Printf("bytes: %v\n", data)
//
//	err = bson2.Unmarshal(data, &v)
//	require.NoError(t, err)
//
//	fmt.Printf("unmarshalled: %v\n", v["data"] == s)
//
//}
//
//func docToBsonBytes(d interface{}) []byte {
//	b, err := bson2.Marshal(d)
//	if err != nil {
//		panic(err)
//	}
//	return b
//}
//
//func docToJsonBytes(d interface{}) []byte {
//	b, err := json.Marshal(d)
//	if err != nil {
//		panic(err)
//	}
//	return b
//}

// BenchmarkBSONv1vs2Comparison runs both v1 and v2 benchmarks for direct comparison
func BenchmarkBSONv1vs2Comparison(b *testing.B) {
	type BusinessUsecaseData struct {
		// Primitive types
		IntType     int
		Int8Type    int8
		Int16Type   int16
		Int32Type   int32
		Int64Type   int64
		UintType    uint
		Uint8Type   uint8
		Uint16Type  uint16
		Uint32Type  uint32
		Uint64Type  uint64
		Float32Type float32
		Float64Type float64
		BoolType    bool
		StringType  string
		ByteType    byte // Alias for uint8
		RuneType    rune // Alias for int32

		// Composite types
		ArrayType  [3]int
		SliceType  []string
		MapType    map[string]int
		StructType struct{ A int }

		// Reference types
		PointerType   *int
		InterfaceType interface{}

		// Others
		EmptyStruct struct{}
	}

	i := 42
	originalData := BusinessUsecaseData{
		IntType:     1,
		Int8Type:    2,
		Int16Type:   3,
		Int32Type:   4,
		Int64Type:   5,
		UintType:    6,
		Uint8Type:   7,
		Uint16Type:  8,
		Uint32Type:  9,
		Uint64Type:  10,
		Float32Type: 12.34,
		Float64Type: 56.78,
		BoolType:    true,
		StringType:  "hello",
		ByteType:    'a',
		RuneType:    '字',

		//ArrayType:  [3]int{1, 2, 3},
		//SliceType:  []string{"a", "b"},
		//MapType:    map[string]int{"key": 1},
		//StructType: struct{ A int }{A: 10},

		PointerType: &i,
		//InterfaceType: 100,

		//EmptyStruct: struct{}{},
	}

	// Marshal data using BSON v1
	var err error
	BSONV1Data, err := bson1.Marshal(originalData)
	if err != nil {
		log.Fatalf("Failed to marshal data using BSON v1: %v", err)
	}

	// Marshal data using BSON v2
	BSONV2Data, err := bson2.Marshal(originalData)
	if err != nil {
		log.Fatalf("Failed to marshal data using BSON v2: %v", err)
	}

	b.Run("BSON_v1", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			var result BusinessUsecaseData
			err := bson1.Unmarshal(BSONV1Data, &result)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("BSON_v2", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			var result BusinessUsecaseData
			err := bson2.Unmarshal(BSONV2Data, &result)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}
