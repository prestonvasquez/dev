package v1

import (
	"context"
	"fmt"
	"log"
	"os"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const testIdxName = "test_vector_index"

// createTestSearchIndexColl returns creates and returns "vstore" collection
// for the "db" database.
func createTestSearchIndexColl(ctx context.Context) (*mongo.Collection, func(context.Context), error) {
	cleanup := func(context.Context) {}

	uri := os.Getenv("MONGODB_URI")
	if uri == "" {
		return nil, cleanup, fmt.Errorf("MONGOD_URI must be set as an Atlas cluster")
	}

	opts := options.Client().ApplyURI(uri)

	client, err := mongo.Connect(ctx, opts)
	if err != nil {
		return nil, cleanup, fmt.Errorf("failed to connect: %w", err)
	}

	cleanup = func(ctx context.Context) { client.Disconnect(ctx) }

	const (
		dbName   = "db"
		collName = "vstore"
	)

	// Create the vectorstore collection
	err = client.Database(dbName).CreateCollection(ctx, collName)
	if err != nil {
		return nil, cleanup, fmt.Errorf("failed to create collection: %w", err)
	}

	coll := client.Database(dbName).Collection(collName)

	return coll, cleanup, nil
}

// From the design rational:
//
// > drivers must not automatically convert this type into a native type by
// > default.
func Example_Decimal128_GODRIVER_3296() {
	opts := options.Client().ApplyURI("mongodb://localhost:27017")

	client, err := mongo.Connect(context.Background(), opts)
	if err != nil {
		log.Fatalf("failed to construct client: %v", err)
	}

	defer client.Disconnect(context.Background())

	coll := client.Database("db").Collection("coll")
	defer coll.Drop(context.Background())

	decs := []interface{}{
		bson.D{{"val", 1.8}},
		bson.D{{"val", 9}},
	}

	if _, err := coll.InsertMany(context.Background(), decs); err != nil {
		log.Fatalf("failed to insert data: %v", err)
	}

	cursor, err := coll.Aggregate(context.Background(), []bson.M{
		{"$group": bson.M{
			"_id":  nil,
			"test": bson.M{"$sum": bson.D{{"$toDecimal", bson.D{{"$ifNull", bson.A{"$dne", "$val"}}}}}},
		}},
	})

	if err != nil {
		log.Fatalf("failed to aggregate: %v", err)
	}

	var result []struct {
		Test primitive.Decimal128 `bson:"test,omitzero"`
	}

	if err := cursor.All(context.Background(), &result); err != nil {
		log.Fatalf("failed to decode cursor: %v", err)
	}

	fmt.Println(result[0].Test.String())
	// Output: 10.80000000000000
}

func ExampleSearchVectors() {
	found := runVectorSearchExample(context.TODO(), "test", []string{"test"})
	//fmt.Println(len(found), scores[0], math.Abs(scores[1]-0.0039) < 1e-3)
	fmt.Println(len(found))
	// Output: 1
}
