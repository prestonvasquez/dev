package main

import (
	"log"

	"go.mongodb.org/mongo-driver/v2/bson"
)

func makeInvalidUTF8() string {
	// 0xff and 0xfe are invalid in UTF-8 sequences.
	return string([]byte{0xff, 0xfe, 0xfd})
}

func main() {
	doc := bson.D{
		{Key: "valid", Value: "hello"},
		{Key: "invalid", Value: makeInvalidUTF8()},
	}

	_, err := bson.Marshal(doc)
	if err != nil {
		log.Fatalf("failed to marshal document: %v", err)
	}
}
