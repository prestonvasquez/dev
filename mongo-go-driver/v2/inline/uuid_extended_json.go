package main

import (
	"bytes"
	"fmt"
	"reflect"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type CustomUUIDEncoder struct{}

func (enc *CustomUUIDEncoder) EncodeValue(_ bson.EncodeContext, vw bson.ValueWriter, v reflect.Value) error {
	// TODO: Logic for encoding uuid
	return nil
}

func main() {
	type SomeStruct struct {
		UUID         uuid.UUID
		UUIDAsString string
	}

	id := uuid.New()

	someStructInst := &SomeStruct{UUID: id, UUIDAsString: id.String()}

	registry := bson.NewRegistry()
	registry.RegisterTypeEncoder(reflect.TypeOf(uuid.UUID{}), &CustomUUIDEncoder{})

	buf := new(bytes.Buffer)
	vw := bson.NewDocumentWriter(buf)

	enc := bson.NewEncoder(vw)

	enc.SetRegistry(registry)

	err := enc.Encode(someStructInst)
	if err != nil {
		panic(err)
	}

	fmt.Println(bson.Raw(buf.Bytes()).String())
}
