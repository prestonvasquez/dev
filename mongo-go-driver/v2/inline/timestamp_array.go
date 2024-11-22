package main

import (
	"fmt"

	"go.mongodb.org/mongo-driver/v2/bson"
)

func main() {
	jsonStr := `[{"$timestamp": {"t": {"$numberInt": "0"}, "i": {"$numberInt": "0"}}}]`
	var arr bson.A
	err := bson.UnmarshalExtJSON([]byte(jsonStr), true, &arr)
	if err != nil {
		panic(err)
	}
	fmt.Println(arr)
}
