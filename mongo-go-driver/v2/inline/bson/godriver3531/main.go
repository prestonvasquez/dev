package main

import (
	"fmt"
	"log"

	"go.mongodb.org/mongo-driver/v2/bson"
)

func main() {
	log.SetFlags(0)

	//v1 := bson.D{{"v", bson.NewDecimal128(3472837370128118064, 3472329395739373616)}}

	d128 := bson.NewDecimal128(3472837370128118064, 3472329395739373616)

	_, err := bson.ParseDecimal128(d128.String())
	fmt.Println("err: ", err)

	//// 1. Raw BSON <-> Decimal128
	//bytes, err := bson.Marshal(v1)
	//if err != nil {
	//	log.Fatalf("failed to marshal: %v", err)
	//}

	//var v2 bson.D
	//if err := bson.Unmarshal(bytes, &v2); err != nil {
	//	log.Fatalf("failed to unmarshal: %v", err)
	//}

	//fmt.Println("raw bson:", v2)

	bytes, err := bson.MarshalExtJSON(v1, true, true)
	if err != nil {
		log.Fatalf("failed to marshal: %v", err)
	}

	var v2 bson.D
	err = bson.UnmarshalExtJSON(bytes, true, &v2)
	if err != nil {
		log.Fatalf("failed to unmarshal: %v", err)
	}

	fmt.Println("ext json:", v2)
}
