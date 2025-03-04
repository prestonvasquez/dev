package main

import (
	"fmt"
	"reflect"
	"sync"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type Registry struct {
	typeEncoders *sync.Map
}

type Encoder interface {
	EncodeValue(vw bson.ValueWriter, val any) error
}

// ValueEncoderFunc is an adapter function that allows a function with the correct signature to be
// used as a ValueEncoder.
type ValueEncoderFunc func(bson.ValueWriter, any) error

// EncodeValue implements the ValueEncoder interface.
func (fn ValueEncoderFunc) EncodeValue(vw bson.ValueWriter, val any) error {
	return fn(vw, val)
}

func nullEncodeValue(vw bson.ValueWriter, val any) error {
	if _, ok := val.(bson.Null); !ok {
		return bson.ValueEncoderError{Name: "NullEncodeValue", Types: []reflect.Type{tNull}, Received: reflect.ValueOf(val)}
	}

	return vw.WriteNull()
}

var tNull = reflect.TypeOf(bson.Null{})

func getTypeEncoderKey(val any) (any, error) {
	switch val.(type) {
	case bson.Null:
		return tNull, nil
	}

	return nil, fmt.Errorf("no default encoder for type %T", val)
}

func LookupEncoder(r *Registry, val any) (Encoder, error) {

	// First check if the user has defined a typeEncoder for the type, if so we need to
	// use that value.
	if r.typeEncoders != nil {
		t, _ := getTypeEncoderKey(val)
		enc, _ := r.typeEncoders.Load(t)

		return enc.(Encoder), nil
	}

	switch val.(type) {
	case bson.Null:
		return ValueEncoderFunc(nullEncodeValue), nil
	}

	return nil, fmt.Errorf("encoder for type %T not found", val)
}

type sliceCodecOptions struct {
	encodeNilAsEmpty bool
}

func newSliceValueEncoder(bson.ValueWriter, sliceCodecOptions) ValueEncoderFunc {
	return func(vw bson.ValueWriter, t any) error {
		encoder, _ := LookupEncoder(t)
		return nil
	}
}

func main() {
	reg := &Registry{typeEncoders: &sync.Map{}}
	reg.typeEncoders.Store(tNull, func() {})

	key, err := getTypeEncoderKey(bson.Null{})
	if err != nil {
		panic(err)
	}

	_, ok := reg.typeEncoders.Load(key)
	fmt.Println(ok)
	//_, err := LookupEncoder[bson.Null](&Registry{})
	//if err != nil {
	//	panic(err)
	//}

}
