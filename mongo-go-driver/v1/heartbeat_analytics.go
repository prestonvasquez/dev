package main

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"

	"go.mongodb.org/mongo-driver/event"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {

	now := time.Now()

	var heartbeatsStartedCounter atomic.Int64
	clientOpts := options.Client().ApplyURI("mongodb://localhost:27017").
		SetServerMonitor(newServerMonitor(&heartbeatsStartedCounter)).
		SetServerMonitoringMode(options.ServerMonitoringModePoll)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, clientOpts)
	if err != nil {
		panic(err)
	}

	defer client.Disconnect(context.Background())

	ticker := time.Tick(1 * time.Second)
	for {
		fmt.Printf("[%05d] total heartbeats until now: %d\n", int(time.Since(now).Seconds()), heartbeatsStartedCounter.Load())
		<-ticker
	}
}

func newServerMonitor(ai64 *atomic.Int64) *event.ServerMonitor {
	monitor := &event.ServerMonitor{
		ServerHeartbeatStarted: func(_ *event.ServerHeartbeatStartedEvent) {
			ai64.Add(1)
		},
	}

	return monitor
}
