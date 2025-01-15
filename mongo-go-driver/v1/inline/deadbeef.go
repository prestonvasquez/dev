package main

import "go.mongodb.org/mongo-driver/bson/primitive"

func main() {
	_, err := primitive.ObjectIDFromHex("deadbeefdeadbeefdeadbeef")
	if err != nil {
		panic(err)
	}
}
