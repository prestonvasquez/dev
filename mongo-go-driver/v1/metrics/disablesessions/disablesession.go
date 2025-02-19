package disablesessions

//type experimentResult struct {
//	ops        int
//	timeoutOps int
//}

//type experimentFn func(ctx context.Context, coll *mongo.Collection) experimentResult
//
//func run(experiment experimentFn) error {
//	// MongoDB connection URI
//	uri := os.Getenv("MONGODB_URI")
//	if uri == "" {
//		uri = "mongodb://localhost:27017"
//	}
//
//	// Connect to MongoDB
//	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(uri))
//	if err != nil {
//		return fmt.Errorf("failed to connect to MongoDB: %v", err)
//	}
//
//	defer func() {
//		if err := client.Disconnect(context.Background()); err != nil {
//			log.Fatalf("Failed to disconnect MongoDB client: %v", err)
//		}
//	}()
//
//	db := client.Database("testdb")
//	coll := db.Collection("coll")
//
//	return nil
//}
