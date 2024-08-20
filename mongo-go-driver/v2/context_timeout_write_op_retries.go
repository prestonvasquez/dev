package v2

//
//import (
//	"context"
//	"devv2/util/eventutil"
//	"devv2/util/failpoint"
//	"log"
//	"time"
//
//	"go.mongodb.org/mongo-driver/v2/bson"
//	"go.mongodb.org/mongo-driver/v2/mongo"
//	"go.mongodb.org/mongo-driver/v2/mongo/options"
//)
//
//// Illustrate timeout behavior on operation-level timeouts (context timeouts)
//// behavior when an operation fails.
//
//func main() {
//	const insertCmd = "insert"
//	const errCode = 9001 // SocketException
//
//	// Create a client that logs insert commands
//	opts := options.Client().ApplyURI("mongodb://localhost:27017").
//		SetMonitor(eventutil.CommandMonitorByName(log.Default(), insertCmd))
//
//	client, err := mongo.Connect(opts)
//	if err != nil {
//		log.Fatalf("failed to connect: %v", err)
//	}
//
//	defer client.Disconnect(context.Background())
//
//	// Create a failpoint that will fail 100x on insert.
//	closeFP, err := failpoint.NewErrorN(context.Background(), client, insertCmd, errCode 100)
//	if err != nil {
//		log.Fatalf("failed to create fail point: %v", err)
//	}
//
//	defer closeFP()
//
//	// Create a context with a timeout of 5 seconds
//	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
//	defer cancel()
//
//	// Run the insert to observe the behavior of a constantly failing operation
//	// with a context timeout.
//	coll := client.Database("db").Collection("coll")
//
//	if _, err := coll.InsertOne(ctx, bson.D{{"x", 1}}); err != nil {
//		log.Fatalf("failed to insert data: %v", err)
//	}
//}
