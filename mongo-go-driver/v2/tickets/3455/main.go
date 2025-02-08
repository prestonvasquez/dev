package main

import (
	"bytes"
	"reflect"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type MyStruct struct {
	TS bson.Timestamp
}

var tTimestamp = reflect.TypeOf(bson.Timestamp{})

func main() {
	ms := &MyStruct{TS: bson.Timestamp{I: 1, T: 2}}

	_, err := bson.Marshal(ms)
	if err != nil {
		panic(err)
	}

	// Create a custom registry.
	buf := new(bytes.Buffer)
	vw := bson.NewDocumentWriter(buf)

	enc := bson.NewEncoder(vw)

	reg := bson.NewRegistry()
	//reg.RegisterTypeEncoder(tTimestamp, bson.ValueEncoderFunc(timestampEncodeValue))

	enc.SetRegistry(reg)

	err = enc.Encode(ms)
	if err != nil {
		panic(err)
	}
}

func timestampEncodeValue(_ bson.EncodeContext, vw bson.ValueWriter, val reflect.Value) error {
	return nil
}
