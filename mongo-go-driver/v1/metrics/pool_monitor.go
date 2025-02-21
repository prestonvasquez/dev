package metrics

import (
	"sync"
	"sync/atomic"
	"time"

	"go.mongodb.org/mongo-driver/event"
)

type PoolMonitor struct {
	*event.PoolMonitor

	connReadyDurMu sync.Mutex
	ConnReadyDur   []float64

	connClosedMu      sync.Mutex
	ConnClosedErrors  map[string]int32
	ConnClosedReasons map[string]int32

	ConnClosed               atomic.Int32
	ConnPendingReadFailed    atomic.Int32
	ConnPendingReadSucceeded atomic.Int32
}

type PoolMonitorCallback func() *PoolMonitor

var _ PoolMonitorCallback = newExpFuncPoolMonitor

func newExpFuncPoolMonitor() *PoolMonitor {
	monitor := &PoolMonitor{
		connReadyDurMu: sync.Mutex{},

		connClosedMu:      sync.Mutex{},
		ConnClosedErrors:  make(map[string]int32),
		ConnClosedReasons: make(map[string]int32),
	}

	monitor.PoolMonitor = &event.PoolMonitor{
		Event: func(pe *event.PoolEvent) {
			switch pe.Type {
			case event.ConnectionClosed:
				monitor.ConnClosed.Add(1)
				monitor.connReadyDurMu.Lock()
				if pe.Error != nil {
					monitor.ConnClosedErrors[pe.Error.Error()]++
				}
				if pe.Reason != "" {
					monitor.ConnClosedReasons[pe.Reason]++
				}
				monitor.connReadyDurMu.Unlock()

			case event.ConnectionReady:
				monitor.connReadyDurMu.Lock()
				monitor.ConnReadyDur = append(monitor.ConnReadyDur, float64(pe.Duration)/float64(time.Millisecond))
				monitor.connReadyDurMu.Unlock()
			}
		},
	}

	return monitor
}
