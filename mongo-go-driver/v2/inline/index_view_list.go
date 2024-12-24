package main

import (
	"context"
	"fmt"
	"log"

	"go.mongodb.org/mongo-driver/v2/mongo"
)

type Collation struct {
	Locale          string `bson:"locale" json:"locale"`
	CaseLevel       bool   `bson:"caseLevel" json:"caseLevel"`
	CaseFirst       string `bson:"caseFirst" json:"caseFirst"`
	Strength        int    `bson:"strength" json:"strength"`
	NumericOrdering bool   `bson:"numericOrdering" json:"numericOrdering"`
	Alternate       string `bson:"alternate" json:"alternate"`
	MaxVariable     string `bson:"maxVariable" json:"maxVariable"`
	Backwards       bool   `bson:"backwards" json:"backwards"`
}

type CustomIndexSpecification struct {
	Collation Collation
}

func main() {
	// Connect to MongoDB
	client, err := mongo.Connect()
	if err != nil {
		log.Fatal(err)
	}

	// Select the database and collection
	database := client.Database("db")
	collection := database.Collection("coll")

	// Get the IndexView for the collection
	indexView := collection.Indexes()

	// List the indexes
	cursor, err := indexView.List(context.TODO())
	if err != nil {
		log.Fatal(err)
	}
	defer cursor.Close(context.TODO())

	// Iterate over the cursor to get index details
	var indexes []CustomIndexSpecification
	if err = cursor.All(context.TODO(), &indexes); err != nil {
		log.Fatal(err)
	}

	// Print the index details
	for _, index := range indexes {
		fmt.Println(index)
	}

	// Disconnect from MongoDB
	if err = client.Disconnect(context.TODO()); err != nil {
		log.Fatal(err)
	}

	fmt.Println("Disconnected from MongoDB!")
}
