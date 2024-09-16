package main

import (
	"bytes"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/bsonrw"
)

func main() {
	buf := new(bytes.Buffer)
	rw, err := bsonrw.NewBSONValueWriter(buf)
	if err != nil {
		panic(err)
	}

	enc, err := bson.NewEncoder(rw)
	if err != nil {
		panic(err)
	}

	enc.OmitZeroStruct()

	type someStruct struct {
		Hi string `bson:"hi"`
	}

	s := someStruct{}

	err = enc.Encode(s)
	if err != nil {
		panic(err)
	}

	// Print the BSON document as Extended JSON by converting it to bson.Raw.

	//bytes, err := bson.Marshal(s)
	//if err != nil {
	//	panic(err)
	//}

	m := map[string]interface{}{}
	if err := bson.Unmarshal(buf.Bytes(), &m); err != nil {
		panic(err)
	}
	//if err := decodeOmitEmpty(bytes, &m); err != nil {
	//	panic(err)
	//}

	fmt.Println(m)
}

func decodeOmitEmpty(bytes []byte, val any) error {
	dec, err := bson.NewDecoder(bsonrw.NewBSONDocumentReader(bytes))
	if err != nil {
		return err
	}

	dec.ZeroMaps()
	dec.ZeroStructs()

	return dec.Decode(val)
}

func encodeOmitEmpty(val any) (*bson.Encoder, error) {
	buf := new(bytes.Buffer)
	rw, err := bsonrw.NewBSONValueWriter(buf)
	if err != nil {
		panic(err)
	}

	enc, err := bson.NewEncoder(rw)

	return enc, err
}
