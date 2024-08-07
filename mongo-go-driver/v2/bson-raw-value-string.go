package main

import (
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/x/bsonx/bsoncore"
)

func main() {
	str := "my string"

	rawValueStr := bson.RawValue{
		Value: bsoncore.AppendString(nil, str),
		Type:  bson.TypeString,
	}

	fmt.Println("val:", rawValueStr.String())
}
