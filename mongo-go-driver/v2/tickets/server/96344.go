package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"reflect"
	"time"
	"unsafe"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/event"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/x/mongo/driver"
	"go.mongodb.org/mongo-driver/v2/x/mongo/driver/description"
	"go.mongodb.org/mongo-driver/v2/x/mongo/driver/mnet"
	"go.mongodb.org/mongo-driver/v2/x/mongo/driver/topology"
)

type CompositeSelector struct {
	Selectors []description.ServerSelector
}

var _ description.ServerSelector = &CompositeSelector{}

// SelectServer combines multiple selectors into a single selector.
func (selector *CompositeSelector) SelectServer(
	topo description.Topology,
	candidates []description.Server,
) ([]description.Server, error) {
	var err error
	for _, sel := range selector.Selectors {
		candidates, err = sel.SelectServer(topo, candidates)
		if err != nil {
			return nil, err
		}
	}

	return candidates, nil
}

func createBlockFP(client *mongo.Client, cmd string, blockTime, iter int) (func(), error) {
	admindb := client.Database("admin")

	// Create a document for the run command that sets a fail command that is always on.
	failCommand := bson.D{
		{Key: "configureFailPoint", Value: "failCommand"},
		{Key: "mode", Value: bson.D{{"times", iter}}},
		{Key: "data", Value: bson.D{
			{Key: "blockConnection", Value: true},
			{Key: "blockTimeMS", Value: blockTime},
			{Key: "failCommands", Value: bson.A{cmd}},
		}},
	}

	err := admindb.RunCommand(context.Background(), failCommand).Err()
	if err != nil {
		return func() {}, err
	}

	return func() {
		doc := bson.D{
			{Key: "configureFailPoint", Value: "failCommand"},
			{Key: "mode", Value: "off"},
		}

		err = admindb.RunCommand(context.Background(), doc).Err()
		if err != nil {
			log.Fatalf("could not disable fail point command: %v", err)
		}
	}, nil
}

func extractTopology(c *mongo.Client) *topology.Topology {
	e := reflect.ValueOf(c).Elem()
	d := e.FieldByName("deployment")
	d = reflect.NewAt(d.Type(), unsafe.Pointer(d.UnsafeAddr())).Elem() // #nosec G103
	return d.Interface().(*topology.Topology)
}

func selectServer(ctx context.Context, c *mongo.Client) (driver.Server, error) {
	return extractTopology(c).SelectServer(ctx, &CompositeSelector{})
}

func checkoutConnection(ctx context.Context, server driver.Server) (*mnet.Connection, error) {
	conn, err := server.Connection(ctx)
	if err != nil {
		return nil, err
	}

	return conn, nil
}

func main() {
	var connectionID string
	commandMonitor := &event.CommandMonitor{
		Started: func(ctx context.Context, cse *event.CommandStartedEvent) {
			if cse.CommandName == "insert" {
				connectionID = cse.ConnectionID
			}
		},
	}

	// Set up MongoDB client options
	opts := options.Client().SetMaxPoolSize(1).SetMonitor(commandMonitor)

	client, err := mongo.Connect(opts)
	if err != nil {
		log.Fatalf("failed to connect to server: %v", err)
	}
	defer func() {
		_ = client.Disconnect(context.Background())
	}()

	// Create a failpoint that will block 1 time for 750ms.
	teardown, err := createBlockFP(client, "insert", 750, 1)
	if err != nil {
		log.Fatalf("failed to create blocking failpoint: %v", err)
	}
	defer teardown()

	coll := client.Database("db").Collection("coll")

	// Attempt to insert with a short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	_, err = coll.InsertOne(ctx, bson.D{})
	if errors.Is(err, context.DeadlineExceeded) {
		log.Printf("Insert failed as expected: %v", err)
	}

	srv, err := selectServer(context.Background(), client)
	if err != nil {
		panic(err)
	}

	conn, err := checkoutConnection(context.Background(), srv)
	if err != nil {
		panic(err)
	}

	fmt.Println(conn.ID(), connectionID)

	//// Attempt to insert with a longer timeout
	//ctx, cancel = context.WithTimeout(context.Background(), 500*time.Millisecond)
	//defer cancel()

	//_, err = coll.InsertOne(ctx, bson.D{})
	//if errors.Is(err, context.DeadlineExceeded) {
	//	log.Printf("Insert failed as expected: %v", err)
	//} else {
	//	log.Println("Insert succeeded unexpectedly")
	//}
}
