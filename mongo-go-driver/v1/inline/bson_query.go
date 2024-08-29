//go:build i1

package main

import (
	"fmt"
	"log"

	"go.mongodb.org/mongo-driver/bson"
)

func main() {
	byteLiteral := []byte(`{x:1}`)

	byteLiteralRes := bson.D{}
	if err := bson.Unmarshal(byteLiteral, &byteLiteralRes); err != nil {
		log.Fatalf("failed to unmarshal bson: %v", err)
	}

	fmt.Println(byteLiteralRes)
}
