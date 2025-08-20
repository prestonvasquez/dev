package main

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/v2/event"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func TestSdamOnClientConstruction(t *testing.T) {
	sdamMonitor := &event.ServerMonitor{
		ServerDescriptionChanged:   func(*event.ServerDescriptionChangedEvent) { fmt.Println("Server description changed") },
		ServerOpening:              func(*event.ServerOpeningEvent) { fmt.Println("Server opening") },
		ServerClosed:               func(*event.ServerClosedEvent) { fmt.Println("Server closed") },
		TopologyDescriptionChanged: func(*event.TopologyDescriptionChangedEvent) { fmt.Println("Topology description changed") },
		TopologyOpening:            func(*event.TopologyOpeningEvent) { fmt.Println("Topology opening") },
		TopologyClosed:             func(*event.TopologyClosedEvent) { fmt.Println("Topology closed") },
		ServerHeartbeatStarted:     func(*event.ServerHeartbeatStartedEvent) { fmt.Println("Server heartbeat started") },
		ServerHeartbeatSucceeded:   func(*event.ServerHeartbeatSucceededEvent) { fmt.Println("Server heartbeat succeeded") },
		ServerHeartbeatFailed:      func(*event.ServerHeartbeatFailedEvent) { fmt.Println("Server heartbeat failed") },
	}

	commandMonitor := &event.CommandMonitor{
		Started: func(_ context.Context, evt *event.CommandStartedEvent) { fmt.Println("Command started:", evt.Command) },
	}

	poolMonitor := &event.PoolMonitor{
		Event: func(evt *event.PoolEvent) {
			fmt.Printf("Pool event: %s, ConnectionID: %d, Address: %s\n",
				evt.Type, evt.ConnectionID, evt.Address)
		},
	}

	clientOpts := options.Client().SetServerMonitor(sdamMonitor).SetMonitor(commandMonitor).SetTimeout(1).SetPoolMonitor(poolMonitor)

	_, err := mongo.Connect(clientOpts)
	require.NoError(t, err, "Failed to connect to MongoDB")

	time.Sleep(1 * time.Second)
}
