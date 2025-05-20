package bson

//package main
//
//import (
//	"context"
//
//	"go.mongodb.org/mongo-driver/v2/bson"
//	"go.mongodb.org/mongo-driver/v2/mongo"
//)
//
//func main() {
//	client, err := mongo.Connect()
//	if err != nil {
//		panic(err)
//	}
//
//	defer client.Disconnect(context.Background())
//
//	intVec := bson.NewVector([]int8{1, 2, 3})
//	doc := bson.D{{"some_int_array", intVec}}
//
//	coll := client.Database("test").Collection("test")
//	coll.InsertOne(context.Background(), doc)
//}

//bytes := make([]byte, 128)

//v, err := bson.NewPackedBitVector(bytes, 0)
//if err != nil {
//	panic(err)
//}

//pb, padding, ok := v.PackedBitOK()
//fmt.Println(len(pb), padding, ok)
