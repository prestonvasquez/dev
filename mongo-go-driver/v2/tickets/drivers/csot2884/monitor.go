package csot2884

import (
	"context"
	"log"
	"sync"

	"go.mongodb.org/mongo-driver/v2/event"
)

type monitor struct {
	commandMonitor *event.CommandMonitor
	poolMonitor    *event.PoolMonitor

	commandStarted   map[string]*event.CommandStartedEvent // cmd -> event
	connectionClosed map[int64]*event.PoolEvent            // connectionID -> event

	eventMu sync.Mutex
}

func newMonitor(shouldLog bool, cmds ...string) *monitor {
	monitor := &monitor{}
	monitor.Reset()

	monitor.commandMonitor = &event.CommandMonitor{
		Started: func(ctx context.Context, cse *event.CommandStartedEvent) {
			for _, cmd := range cmds {
				if cse.CommandName == cmd {
					if shouldLog {
						log.Printf("command started: %+v\n", cse)
					}

					monitor.eventMu.Lock()
					monitor.commandStarted[cmd] = cse
					monitor.eventMu.Unlock()
				}
			}
		},
		Failed: func(ctx context.Context, cse *event.CommandFailedEvent) {
			for _, cmd := range cmds {
				if cse.CommandName == cmd {
					if shouldLog {
						log.Printf("command failed: %+v\n", cse)
					}
				}
			}
		},
	}

	monitor.poolMonitor = &event.PoolMonitor{
		Event: func(pe *event.PoolEvent) {
			switch pe.Type {
			case event.ConnectionCheckedIn:
				if shouldLog {
					log.Printf("connection checked in: %+v\n", pe)
				}
			case event.ConnectionCheckedOut:
				if shouldLog {
					log.Printf("connection checked out: %+v\n", pe)
				}
			case event.ConnectionClosed:
				if shouldLog {
					log.Printf("connection closed: %+v\n", pe)
				}

				monitor.eventMu.Lock()
				monitor.connectionClosed[pe.ConnectionID] = pe
				monitor.eventMu.Unlock()
			}
		},
	}

	return monitor
}

func (m *monitor) Reset() {
	m.commandStarted = map[string]*event.CommandStartedEvent{}
	m.connectionClosed = map[int64]*event.PoolEvent{}
	m.eventMu = sync.Mutex{}
}
