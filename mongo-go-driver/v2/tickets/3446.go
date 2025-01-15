package main

import (
	"context"
	"fmt"
	"log"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type User struct {
	Email string `bson:"email"`
}

func main() {
	// Connect to MongoDB
	client, err := mongo.Connect()
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

	//// Insert a new document
	//newUser := bson.D{{Key: "email", Value: "testuser@example.com"}}
	//insertResult, err := collection.InsertOne(context.Background(), newUser)
	//if err != nil {
	//	fmt.Println("Error inserting document:", err)
	//	return
	//}

	//// Use the inserted ID for the lookup
	//uid := insertResult.InsertedID.(bson.ObjectID)
	//fmt.Println("Inserted document ID:", uid)

	uid, err := bson.ObjectIDFromHex("67632369f1526917193ee21c")

	var user User
	opts := options.FindOne().SetProjection(map[string]any{"email": 1})

	// Find one document using the inserted document's ID
	err = collection.FindOne(context.Background(), bson.M{"_id": uid}, opts).Decode(&user)
	if err != nil {
		fmt.Println("Error finding user:", err)
		return
	}

	fmt.Println("User:", user)
}
