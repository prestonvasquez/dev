package main

import (
	"fmt"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type zeroer struct {
	MyString string
}

//func (zeroer) IsZero() bool {
//	return true
//}

//var _ bson.Zeroer = zeroer{}

type myStruct struct {
	Zero zeroer `bson:",omitzero"`
	//I64     int64          `bson:",omitempty"`
	//F64     float64        `bson:",omitempty"`
	//String  string         `bson:",omitempty"`
	//Boolean bool           `bson:",omitempty"`
	//Array   []int          `bson:",omitempty"`
	//Map     map[string]int `bson:",omitempty"`
	//Bytes   []byte         `bson:",omitempty"`
	//Time    time.Time      `bson:",omitempty"`
	//Pointer *int           `bson:",omitempty"`
}

func main() {
	s := myStruct{}
	bytes, err := bson.Marshal(s)
	if err != nil {
		panic(err)
	}

	fmt.Println(bytes) // [5,0,0,0,0]
}
