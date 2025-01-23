package main

//import (
//	"context"
//	"errors"
//	"fmt"
//	"log"
//	"os"
//	"runtime"
//	"sort"
//	"sync/atomic"
//	"time"
//
//	"github.com/google/uuid"
//	"go.mongodb.org/mongo-driver/v2/bson"
//	"go.mongodb.org/mongo-driver/v2/event"
//	"go.mongodb.org/mongo-driver/v2/mongo"
//	"go.mongodb.org/mongo-driver/v2/mongo/options"
//	"golang.org/x/exp/rand"
//	"gonum.org/v1/gonum/stat"
//)
//
//const defaultConnectionChurnDB = "connectionChurnDB"
//
//func main() {
//	uri := os.Getenv("MONGODB_URI")
//	if uri == "" {
//		uri = "mongodb://localhost:27017"
//	}
//
//	connCreateTimes := []float64{}
//
//	var connectionsClosed int32
//	var connectionsCheckedOut int32
//
//	monitor := &event.PoolMonitor{
//		Event: func(pe *event.PoolEvent) {
//			switch pe.Type {
//			case event.ConnectionReady:
//				connCreateTimes = append(connCreateTimes, float64(pe.Duration)/float64(time.Millisecond))
//			case event.ConnectionClosed:
//				atomic.AddInt32(&connectionsClosed, 1)
//			case event.ConnectionCheckedOut:
//				atomic.AddInt32(&connectionsCheckedOut, 1)
//			}
//		},
//	}
//
//	clientOpts := options.Client().
//		ApplyURI(uri).
//		SetMaxPoolSize(1).SetPoolMonitor(monitor)
//
//	client, err := mongo.Connect(clientOpts)
//	if err != nil {
//		log.Fatalf("failed to establish connection: %v", err)
//	}
//
//	defer func() {
//		if err := client.Disconnect(context.Background()); err != nil {
//			log.Fatalf("failed to disconnect: %v", err)
//		}
//	}()
//
//	collName, err := loadLargeCollection(context.Background(), 1000)
//	if err != nil {
//		log.Fatalf("failed to load large collection: %v", err)
//	}
//
//	coll := client.Database("db").Collection(collName)
//	defer func() {
//		if err := coll.Drop(context.Background()); err != nil {
//			log.Fatalf("failed to drop collection: %v", err)
//		}
//	}()
//
//	for i := 0; i < 100_000; i++ {
//		ctx, cancel := context.WithTimeout(context.Background(), 150*time.Millisecond)
//		defer cancel()
//
//		err := coll.FindOne(ctx, bson.D{}).Err()
//		if !errors.Is(err, context.DeadlineExceeded) {
//			//log.Fatalf("failed to apply timeout")
//		}
//	}
//
//	fmt.Println(connCreateTimes)
//	fmt.Println("connections closed: ", connectionsClosed)
//	fmt.Println("connections checked out: ", connectionsCheckedOut)
//
//	sort.Float64s(connCreateTimes)
//	fmt.Println("median conn establish time (ms): ", stat.Quantile(0.5, stat.Empirical, connCreateTimes, nil))
//}
//
//func createFPAlwaysOn(ctx context.Context, client *mongo.Client, blockTime int) (func(), error) {
//	admindb := client.Database("admin")
//
//	// Create a document for the run command that sets a fail command that is always on.
//	failCommand := bson.D{
//		{Key: "configureFailPoint", Value: "failCommand"},
//		{Key: "mode", Value: "alwaysOn"},
//		{Key: "data", Value: bson.D{
//			{Key: "blockConnection", Value: true},
//			{Key: "blockTimeMS", Value: blockTime},
//			{Key: "failCommands", Value: bson.A{"find"}}},
//		},
//	}
//
//	err := admindb.RunCommand(ctx, failCommand).Err()
//	if err != nil {
//		return func() {}, err
//	}
//
//	return func() {
//		doc := bson.D{
//			{Key: "configureFailPoint", Value: "failCommand"},
//			{Key: "mode", Value: "off"},
//		}
//
//		err = admindb.RunCommand(ctx, doc).Err()
//		if err != nil {
//			log.Fatalf("could not disable fail point command: %v", err)
//		}
//	}, nil
//}
//
//// loadLargeCollection will dedicate a worker pool to inserting test data into
//// an unindexed collection. Each record is 31 bytes in size.
//func loadLargeCollection(ctx context.Context, size int, opts ...*options.ClientOptions) (string, error) {
//	client, err := mongo.Connect(opts...)
//	if err != nil {
//		return "", fmt.Errorf("failed to create client: %w", err)
//	}
//
//	defer func() {
//		if err := client.Disconnect(ctx); err != nil {
//			panic(err)
//		}
//	}()
//
//	// Initialize a collection with the name "large<uuid>".
//	collName := fmt.Sprintf("large%s", uuid.NewString())
//
//	goRoutines := runtime.NumCPU()
//
//	// Partition the volume into equal sizes per go routine. Use the floor if the
//	// volume is not divisible by the number of goroutines.
//	perGoroutine := size / goRoutines
//
//	docs := make([]interface{}, perGoroutine)
//	for i := range docs {
//		docs[i] = bson.D{
//			{Key: "field1", Value: rand.Int63()},
//			{Key: "field2", Value: rand.Int31()},
//		}
//	}
//
//	errs := make(chan error, goRoutines)
//	done := make(chan struct{}, goRoutines)
//
//	coll := client.Database(defaultConnectionChurnDB).Collection(collName)
//
//	for i := 0; i < int(goRoutines); i++ {
//		go func(i int) {
//			_, err := coll.InsertMany(ctx, docs)
//			if err != nil {
//				errs <- fmt.Errorf("goroutine %v failed: %w", i, err)
//			}
//
//			done <- struct{}{}
//		}(i)
//	}
//
//	go func() {
//		defer close(errs)
//		defer close(done)
//
//		for i := 0; i < int(goRoutines); i++ {
//			<-done
//		}
//	}()
//
//	// Await errors and return the first error encountered.
//	for err := range errs {
//		if err != nil {
//			return "", err
//		}
//	}
//
//	return collName, nil
//}
