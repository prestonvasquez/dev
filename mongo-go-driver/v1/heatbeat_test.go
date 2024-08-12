package v1

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/event"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func newServerMonitor(ai64 *atomic.Int64) *event.ServerMonitor {
	monitor := &event.ServerMonitor{
		ServerHeartbeatStarted: func(_ *event.ServerHeartbeatStartedEvent) {
			ai64.Add(1)
		},
	}

	return monitor
}

func TestHeartbeatAnalytics(t *testing.T) {
	now := time.Now()

	var heartbeatsStartedCounter atomic.Int64
	clientOpts := options.Client().ApplyURI("mongodb://localhost:27017").
		SetServerMonitor(newServerMonitor(&heartbeatsStartedCounter)).
		SetServerMonitoringMode(options.ServerMonitoringModePoll)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, clientOpts)
	assert.NoError(t, err, "failed to connect")

	defer client.Disconnect(context.Background())

	ticker := time.Tick(1 * time.Second)
	for {
		t.Logf("[%05d] total heartbeats until now: %d\n", int(time.Since(now).Seconds()), heartbeatsStartedCounter.Load())
		<-ticker
	}
}
