package main

import (
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
)

func main() {
	jsonStr := `[{"$timestamp": {"t": {"$numberInt": "0"}, "i": {"$numberInt": "0"}}}]`
	//jsonStr := `[{"":{"$timestamp": {"t": {"$numberInt": "0"}, "i": {"$numberInt": "0"}}}}]`
	//jsonStr := `{"":{"$timestamp":{"t":0,"i":-1}}}`

	var arr bson.A
	//var arr bson.D
	err := bson.UnmarshalExtJSON([]byte(jsonStr), true, &arr)
	if err != nil {
		panic(err)
	}
	fmt.Println(arr)
}
