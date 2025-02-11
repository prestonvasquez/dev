package main

import (
	"fmt"
	"runtime/debug"

	"go.mongodb.org/mongo-driver/bson"
)

func main() {
	var v struct {
		T myType `bson:"t"`
	}

	err := bson.Unmarshal(docToBytes(bson.D{{"t", nil}, {"i", nil}}), &v)
	if err != nil {
		panic(err)
	}
	fmt.Println(v)

}

type myType struct {
	A string `bson:"a"`
}

func (t *myType) UnmarshalBSON(b []byte) error {
	debug.PrintStack()
	fmt.Println("myType")
	return nil
}

func docToBytes(d interface{}) []byte {
	b, err := bson.Marshal(d)
	if err != nil {
		panic(err)
	}
	return b
}
