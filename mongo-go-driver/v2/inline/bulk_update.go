package main

type Listing struct {
	ID string `bson:"_id"`
}

func main() {
	//listings := []Listing{}

	//models := make([]mongo.WriteModel, len(listings))
	//for i := range models {
	//	models[i] = mongo.NewUpdateOneModel().
	//		SetFilter(bson.D{{"_id", listings[i].ID}}).
	//		SetUpdate(bson.D{{"$set", bson.D{{"embeddings", embeddings[i]}}}})
	//}

	//bulkOpts := options.BulkWrite().SetOrdered(false)

	//result, err := collection.BulkWrite(context.Background(), models, bulkOpts)
	//if err != nil {
	//	log.Fatalf("bulkWrite fialed: %v", err)
	//}
}
