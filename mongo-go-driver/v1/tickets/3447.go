package main

import (
	"context"
	"fmt"
	"log"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type User struct {
	Email string `bson:"email"`
}

func main() {
	// Connect to MongoDB
	client, err := mongo.Connect(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	// Disconnect when the function returns
	defer func() {
		if err = client.Disconnect(context.Background()); err != nil {
			log.Fatal(err)
		}
	}()

	// Get a handle for your collection
	collection := client.Database("your_database").Collection("users")

	uid, err := primitive.ObjectIDFromHex("deadbeefdeadbeefdeadbeef")
	if err != nil {
		panic(err)
	}

	//// Insert a new document to query
	//newUser := bson.D{
	//	{Key: "_id", Value: uid},
	//	{Key: "email", Value: "testuser@example.com"},
	//}

	//_, err = collection.InsertOne(context.Background(), newUser)
	//if err != nil {
	//	fmt.Println("Error inserting document:", err)
	//	return
	//}

	var user User
	opts := options.FindOne().SetProjection(map[string]any{"email": 1})

	// Find one document using the inserted document's ID
	err = collection.FindOne(context.Background(), bson.M{"_id": uid}, opts).Decode(&user)
	if err != nil {
		fmt.Println("Error finding user:", err)
	}

	fmt.Println("User:", user)
}
