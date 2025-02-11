package main

//// Configuration constants
//const (
//	runDuration       = 1 * time.Minute         // Duration to run the test
//	maxAwaitTime      = 2000 * time.Millisecond // Maximum await time for cursors
//	experimentTimeout = 1000 * time.Millisecond // Timeout for experiment queries
//	collectionCount   = 10_000                  // Number of documents in the collection
//	collectionSize    = 1024 * 1024             // Size of the capped collection (1 MB)
//)
//
//func main() {
//	// MongoDB connection URI
//	uri := os.Getenv("MONGODB_URI")
//	if uri == "" {
//		uri = "mongodb://localhost:27017"
//	}
//
//	// Connect to MongoDB
//	client, err := mongo.Connect(options.Client().ApplyURI(uri))
//	if err != nil {
//		log.Fatalf("Failed to connect to MongoDB: %v", err)
//	}
//	defer func() {
//		if err := client.Disconnect(context.Background()); err != nil {
//			log.Fatalf("Failed to disconnect MongoDB client: %v", err)
//		}
//	}()
//
//	// Preload data into a collection
///	collectionName, err := preloadLargeCollection(context.Background(), collectionCount, client)
//	if err != nil {
//		log.Fatalf("Failed to preload collection: %v", err)
//	}
//
//	// Create a cancellable context
//	ctx, cancel := context.WithCancel(context.Background())
//
//	// Run the experiment in a separate goroutine
//	go runTimeoutExperiment(ctx, collectionName)
//
//	// Wait for the experiment duration
//	time.Sleep(runDuration)
//	cancel() // Cancel the context to stop the experiment
//
//	time.Sleep(10 * time.Second) // Wait for cleanup
//}
//
//// runTimeoutExperiment executes queries with a timeout to simulate stress on MongoDB.
//func runTimeoutExperiment(ctx context.Context, collectionName string) {
//	// Get MongoDB URI from environment
//	uri := os.Getenv("MONGODB_URI")
//	if uri == "" {
//		uri = "mongodb://localhost:27017"
//	}
//
//	var connectionsClosed atomic.Int64
//	connectionReadyDurationsMu := sync.Mutex{}
//	connectionReadyDurations := []float64{}
//
//	// Pool monitor to track connection events
//	poolMonitor := &event.PoolMonitor{
//		Event: func(pe *event.PoolEvent) {
//			switch pe.Type {
//			case event.ConnectionClosed:
//				connectionsClosed.Add(1)
//			case event.ConnectionReady:
//				connectionReadyDurationsMu.Lock()
//				connectionReadyDurations = append(connectionReadyDurations, float64(pe.Duration)/float64(time.Millisecond))
//				connectionReadyDurationsMu.Unlock()
//			}
//		},
//	}
//
//	var commandFailed atomic.Int64
//	var commandSucceeded atomic.Int64
//	var commandStarted atomic.Int64
//
//	// Command monitor to track command events
//	cmdMonitor := &event.CommandMonitor{
//		Started: func(_ context.Context, cse *event.CommandStartedEvent) {
//			if cse.CommandName == "getMore" {
//				fmt.Println("command: ", cse.Command)
//				commandStarted.Add(1)
//			}
//		},
//		Succeeded: func(_ context.Context, cse *event.CommandSucceededEvent) {
//			if cse.CommandName == "getMore" {
//				fmt.Println("reply: ", cse.Reply, cse.Duration)
//				commandSucceeded.Add(1)
//			}
//		},
//		Failed: func(_ context.Context, evt *event.CommandFailedEvent) {
//			if evt.CommandName == "getMore" {
//				commandFailed.Add(1)
//			}
//		},
//	}
//
//	connectionPendingReadDurationMu := sync.Mutex{}
//	connectionPendingReadDurations := []float64{}
//	connectionPendingCount := 0
//
//	// Topology callback to track pending read durations
//	topology.BGReadCallback = func(addr string, start, read time.Time, errs []error, connClosed bool) {
//		connectionPendingReadDurationMu.Lock()
//		connectionPendingReadDurations = append(connectionPendingReadDurations, float64(time.Since(start))/float64(time.Millisecond))
//		connectionPendingCount++
//		connectionPendingReadDurationMu.Unlock()
//	}
//
//	// Connect to MongoDB with monitors
//	client, err := mongo.Connect(options.Client().ApplyURI(uri).SetPoolMonitor(poolMonitor).SetMonitor(cmdMonitor).
//		SetHeartbeatInterval(500 * time.Millisecond))
//	if err != nil {
//		log.Fatalf("failed to connect: %v", err)
//	}
//	defer func() {
//		if err := client.Disconnect(context.Background()); err != nil {
//			log.Fatalf("failed to disconnect: %v", err)
//		}
//	}()
//
//	// Wait for minimum RTT to register
//	time.Sleep(1 * time.Second)
//
//	db := client.Database("testdb")
//	coll := db.Collection(collectionName)
//
//	log.Println("[Experiment] starting timeout queries")
//
//	opCount := 0
//	timeoutErrCount := 0
//	opDurs := []float64{}
//
//	for {
//		select {
//		case <-ctx.Done():
//			// Log results of the experiment
//			log.Printf(`[Experiment] results: {
//	"connections_closed": %v,
//	"connections_ready": %v,
//	"commands_failed": %v,
//	"commands_started": %v,
//	"commands_succeeded": %v,
//	"average_connection_ready_duration_ms": %v,
//	"median_connection_ready_duration_ms": %v,
//	"op_count": %v,
//	"timeout_err_count": %v,
//	"average_op_duration": %v,
//	"median_op_duration": %v
//	"pending_read_count": %v,
//	"average_pending_read_dur_ms": %v,
//	"median_pending_read_dur_ms": %v
//}`, connectionsClosed.Load(),
//				len(connectionReadyDurations),
//				commandFailed.Load(),
//				commandStarted.Load(),
//				commandSucceeded.Load(),
//				average(connectionReadyDurations),
//				median(connectionReadyDurations),
//				opCount,
//				timeoutErrCount,
//				average(opDurs),
//				median(opDurs),
//				connectionPendingCount,
//				average(connectionPendingReadDurations),
//				median(connectionPendingReadDurations),
//			)
//			return
//		default:
//			// Create a context with a timeout for each operation
//			ctx, cancel := context.WithTimeout(context.Background(), experimentTimeout)
//			ctx = context.WithValue(ctx, "latency_context", true)
//
//			// Define a query to simulate
//			query := bson.D{}
//
//			// Measure operation duration
//			opStart := time.Now()
//			err := findCursor(ctx, query, coll)
//			opDurs = append(opDurs, float64(time.Since(opStart))/float64(time.Millisecond))
//
//			if errors.Is(err, context.DeadlineExceeded) {
//				timeoutErrCount++
//			}
//
//			opCount++
//			cancel() // Cancel the context
//		}
//	}
//}
//
//// preloadLargeCollection populates a MongoDB collection with random data.
//func preloadLargeCollection(ctx context.Context, size int, client *mongo.Client) (string, error) {
//	//collectionName := fmt.Sprintf("large_%s", uuid.NewString())
//
//	collectionName := "lagecollection"
//	collection := client.Database("testdb").Collection(collectionName)
//
//	if err := setupCappedCollection(context.Background(), collection); err != nil {
//		log.Fatalf("failed to setup a capped collection: %v", err)
//	}
//
//	workerCount := runtime.NumCPU() // Number of worker goroutines
//	batchSize := size / workerCount
//
//	documents := make([]interface{}, batchSize)
//	for i := range documents {
//		documents[i] = bson.D{{Key: "field1", Value: rand.Int63()}, {Key: "field2", Value: rand.Int31()}}
//	}
//
//	errChan := make(chan error, workerCount)
//	doneChan := make(chan struct{}, workerCount)
//
//	// Insert documents using multiple goroutines
//	for i := 0; i < workerCount; i++ {
//		go func() {
//			if _, err := collection.InsertMany(ctx, documents); err != nil {
//				errChan <- err
//			} else {
//				doneChan <- struct{}{}
//			}
//		}()
//	}
//
//	// Wait for all insert operations to complete
//	for i := 0; i < workerCount; i++ {
//		select {
//		case err := <-errChan:
//			return "", err
//		case <-doneChan:
//		}
//	}
//
//	return collectionName, nil
//}
//
//// median calculates the median of a sorted slice of float64 numbers.
//func median(sortedData []float64) float64 {
//	n := len(sortedData)
//	if n == 0 {
//		return 0 // Handle empty slice
//	}
//	if n%2 == 1 {
//		return sortedData[n/2] // Odd number of elements
//	}
//	// Even number of elements
//	mid := n / 2
//	return (sortedData[mid-1] + sortedData[mid]) / 2
//}
//
//// average calculates the mean of a slice of float64 numbers.
//func average(data []float64) float64 {
//	if len(data) == 0 {
//		return 0 // Handle empty slice to avoid division by zero
//	}
//	sum := 0.0
//	for _, value := range data {
//		sum += value
//	}
//	return sum / float64(len(data))
//}
//
//// findCursor executes a find command with a tailable cursor on the collection.
//func findCursor(ctx context.Context, query bson.D, coll *mongo.Collection) error {
//	// Configure the tailable cursor
//	findOptions := options.Find()
//	findOptions.SetCursorType(options.TailableAwait)
//	findOptions.SetBatchSize(1)
//	findOptions.SetMaxAwaitTime(maxAwaitTime)
//
//	// Start watching for changes in the capped collection
//	cursor, err := coll.Find(ctx, query, findOptions)
//	if err != nil {
//		return err
//	}
//	defer cursor.Close(ctx)
//
//	var docs []bson.Raw
//	if err := cursor.All(ctx, &docs); err != nil {
//		fmt.Println("cursor error: ", err)
//		return err
//	} else {
//		fmt.Println("how?")
//	}
//
//	return nil
//}
//
//// Helper function to set up a capped collection (if not already existing)
//func setupCappedCollection(ctx context.Context, collection *mongo.Collection) error {
//	db := collection.Database()
//	collectionName := collection.Name()
//
//	// Check if the collection exists
//	collections, err := db.ListCollectionNames(ctx, bson.M{"name": collectionName})
//	if err != nil {
//		return err
//	}
//
//	if len(collections) == 0 {
//		// Collection does not exist, create it as a capped collection
//		cappedOpts := options.CreateCollection().SetCapped(true).SetSizeInBytes(collectionSize) // 1 MB size
//		err = db.CreateCollection(ctx, collectionName, cappedOpts)
//		if err != nil {
//			return err
//		}
//		log.Println("Capped collection created:", collectionName)
//	} else {
//		log.Println("Capped collection already exists:", collectionName)
//	}
//
//	return nil
//}
