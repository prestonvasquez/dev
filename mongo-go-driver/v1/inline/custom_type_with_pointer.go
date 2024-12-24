package main

import (
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/bsontype"
	"go.mongodb.org/mongo-driver/x/bsonx/bsoncore"
)

type DBInt64 int64

type Product struct {
	TotalForSell *DBInt64 `json:"total_for_sell" bson:"total_for_sell,omitempty"`
}

func (i *DBInt64) UnmarshalBSONValue(t bsontype.Type, value []byte) error {
	return nil
}

func main() {
	idx, doc := bsoncore.AppendDocumentStart(nil)
	doc = bsoncore.AppendNullElement(doc, "total_for_sell")

	doc, err := bsoncore.AppendDocumentEnd(doc, idx)
	if err != nil {
		panic(err)
	}

	bytes := bson.Raw(doc)

	got := Product{}
	if err := bson.Unmarshal(bytes, &got); err != nil {
		panic(err)
	}

	if got.TotalForSell != nil {
		fmt.Println("null value decoded")
	}
}
