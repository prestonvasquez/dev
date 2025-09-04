package session_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/event"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/mongo/readconcern"
	"go.mongodb.org/mongo-driver/v2/mongo/readpref"
	"go.mongodb.org/mongo-driver/v2/mongo/writeconcern"
)

func newCommandMonitor() *event.CommandMonitor {
	return &event.CommandMonitor{
		Started: func(ctx context.Context, evt *event.CommandStartedEvent) {
			fmt.Printf("Command started: %s\n", evt.Command)
		},
		//Succeeded: func(ctx context.Context, evt *event.CommandSucceededEvent) {
		//	fmt.Printf("Command succeeded: %s\n", evt.Command)
		//},
		//Failed: func(ctx context.Context, evt *event.CommandFailedEvent) {
		//	fmt.Printf("Command failed: %s\n", evt.Failure)
		//},
	}
}

func TestSameSessionMultipleCursor(t *testing.T) {
	client, err := mongo.Connect()
	require.NoError(t, err)

	defer client.Disconnect(context.Background())

	sess, err := client.StartSession()
	require.NoError(t, err)

	coll := client.Database("testdb").Collection("tstcoll")
	_ = coll.Drop(context.Background())

	ctx := mongo.NewSessionContext(context.Background(), sess)

	// Seed some data
	for i := 0; i < 10; i++ {
		coll.InsertOne(ctx, bson.D{{"status", "active"}})
		coll.InsertOne(ctx, bson.D{{"status", "inactive"}})
	}

	cur1, err := coll.Find(ctx, bson.D{{"status", "active"}})
	require.NoError(t, err)

	cur2, err := coll.Find(ctx, bson.D{{"status", "inactive"}})
	require.NoError(t, err)

	allActive := []bson.D{}
	require.NoError(t, cur1.All(ctx, &allActive))
	assert.Len(t, allActive, 10)

	allInactive := []bson.D{}
	require.NoError(t, cur2.All(ctx, &allInactive))
	assert.Len(t, allInactive, 10)
}

// (InvalidOptions) writeConcern is not allowed within a multi-statement transaction
func TestInvalidWriteConcernInTransaction(t *testing.T) {
	clientOpts := options.Client().
		SetWriteConcern(writeconcern.Majority()).
		SetReadConcern(readconcern.Majority()).
		SetReadPreference(readpref.PrimaryPreferred()).
		SetMonitor(newCommandMonitor())

	client, err := mongo.Connect(clientOpts)
	require.NoError(t, err)

	defer client.Disconnect(context.Background())

	sessionOpts := options.Session().SetCausalConsistency(true)
	session, err := client.StartSession(sessionOpts)
	require.NoError(t, err)

	defer session.EndSession(context.Background())

	txnOpts := options.Transaction().SetReadPreference(readpref.Primary())
	fmt.Println("Starting transaction with invalid write concern")
	_, err = session.WithTransaction(context.Background(), func(sc context.Context) (interface{}, error) {

		coll := client.Database("your_database_name").Collection("coll")
		res, err := coll.InsertOne(sc, bson.D{{"_id", 0}})
		fmt.Println("done")
		return res, err
		// Create the insert command
		//command := bson.D{
		//	{Key: "insert", Value: "t"},
		//	{Key: "documents", Value: bson.A{bson.D{}}},
		//	{Key: "writeConcern", Value: bson.D{{Key: "w", Value: 1}}},
		//}

		//var result bson.M
		//// Run the command in the session context
		//err := client.Database("your_database_name").RunCommand(sc, command).Decode(&result)
		//if err != nil {
		//	return nil, fmt.Errorf("failed to run command: %w", err)
		//}

		//fmt.Println("Command result:", result)
		//return result, nil
	}, txnOpts)

	require.NoError(t, err)
}

//func TestInvalidWriteConcernInTransaction2(t *testing.T) {
//	clientOpts := options.Client().
//		SetWriteConcern(writeconcern.Majority()).
//		SetReadConcern(readconcern.Majority()).
//		SetReadPreference(readpref.PrimaryPreferred())
//
//	client, err := mongo.Connect(clientOpts)
//	require.NoError(t, err)
//
//	defer client.Disconnect(context.Background())
//
//	ctx := context.Background()
//
//	sessionOpts := options.Session().SetCausalConsistency(true)
//	session, err := client.StartSession(sessionOpts)
//	require.NoError(t, err)
//
//	defer session.EndSession(ctx)
//
//	txnStarted := make(chan struct{})
//
//	go func() {
//		_, err = session.WithTransaction(ctx, func(sc context.Context) (interface{}, error) {
//
//			_, err = client.Database("your_database_name").Collection("coll").InsertOne(
//				context.Background(),
//				bson.D{{"_id", 0}},
//			)
//
//			close(txnStarted)
//			_, err = client.Database("your_database_name").Collection("coll").UpdateByID(
//				ctx,
//				0,
//				bson.D{{"$set", bson.D{{"finishAtTs", primitive.Timestamp{T: 1, I: 0}}}}},
//			)
//			return nil, err
//		})
//
//		time.Sleep(5 * time.Minute)
//	}()
//
//	<-txnStarted
//	sessionCtx := mongo.NewSessionContext(ctx, session)
//	_, err = client.Database("your_database_name").Collection("coll").InsertOne(
//		sessionCtx,
//		bson.D{{"_id", 0}},
//	)
//
//	_, err = client.Database("your_database_name").Collection("coll").UpdateByID(
//		sessionCtx,
//		0,
//		bson.D{{"$set", bson.D{{"finishAtTs", primitive.Timestamp{T: 1, I: 0}}}}},
//	)
//
//	require.NoError(t, err)
//}
//
