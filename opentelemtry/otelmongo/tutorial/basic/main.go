package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.opentelemetry.io/contrib/instrumentation/go.mongodb.org/mongo-driver/mongo/otelmongo"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

func main() {
	// ===== Set up OpenTelemetry Tracer =====
	f, err := os.Create("traces.json")
	if err != nil {
		log.Fatalf("failed to create traces.json: %v", err)
	}

	defer f.Close()

	// Create an stdout exporter to view the spans on the console. In practice
	// this would export to jaeger-ui, etc, for example.
	exporter, err := stdouttrace.New(
		stdouttrace.WithPrettyPrint(),
		stdouttrace.WithWriter(f),
	)
	if err != nil {
		log.Fatalf("failed to create stdout exporter: %v", err)
	}

	// Create a tracer provider with a batch span processor that will send
	// completed span batches to the exporter.
	traceProvider := sdktrace.NewTracerProvider(
		sdktrace.WithSyncer(exporter),
	)

	// Register the tracer provider globally.
	otel.SetTracerProvider(traceProvider)
	tracer := otel.Tracer("example-tracer")

	// ===== Set up MongoDB with otelmongo instrumentation =====

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	clientOpt := options.Client().SetMonitor(otelmongo.NewMonitor())

	client, err := mongo.Connect(ctx, clientOpt)
	if err != nil {
		log.Fatalf("failed to connect to MongoDB: %v", err)
	}

	defer func() {
		if err := client.Disconnect(ctx); err != nil {
			log.Fatalf("failed to disconnect from MongoDB: %v", err)
		}
	}()

	// ===== Start a Trace Span and Perform a MongoDB Operation =====

	ctx, span := tracer.Start(ctx, "mongo-insert-operation")
	defer span.End()

	// Access a collection
	coll := client.Database("exampledb").Collection("examplecoll")

	result, err := coll.InsertOne(ctx, bson.D{{"name", "otelmongo-example"}})
	if err != nil {
		span.RecordError(err)

		log.Fatalf("failed to insert document: %v", err)
	}

	fmt.Printf("Inserted document with ID: %v\n", result.InsertedID)

	// ===== Shutdown OpenTelemetry =====

	if err := exporter.Shutdown(context.Background()); err != nil {
		log.Fatalf("failed to shutdown exporter: %v", err)
	}
}
