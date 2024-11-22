package csot

import (
	"context"
	"log"
	"sync"

	"go.mongodb.org/mongo-driver/v2/event"
)

type monitor struct {
	commandMonitor *event.CommandMonitor
	poolMonitor    *event.PoolMonitor

	commandStarted                 map[string][]*event.CommandStartedEvent // cmd -> event
	commandFailed                  map[string][]*event.CommandFailedEvent
	connectionCheckedOut           map[int64][]*event.PoolEvent
	connectionCheckedIn            map[int64][]*event.PoolEvent
	connectionPendingReadStarted   map[int64][]*event.PoolEvent
	connectionPendingReadSucceeded map[int64][]*event.PoolEvent
	connectionPendingReadFailed    map[int64][]*event.PoolEvent
	connectionClosed               map[int64][]*event.PoolEvent // connectionID -> event

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
					monitor.commandStarted[cmd] = append(monitor.commandStarted[cmd], cse)
					monitor.eventMu.Unlock()
				}
			}
		},
		Succeeded: func(ctx context.Context, cse *event.CommandSucceededEvent) {
			for _, cmd := range cmds {
				if cse.CommandName == cmd {
					if shouldLog {
						log.Printf("command Succeeded: %+v\n", cse)
					}
				}
			}
		},
		Failed: func(ctx context.Context, cse *event.CommandFailedEvent) {
			for _, cmd := range cmds {
				if cse.CommandName == cmd {
					if shouldLog {
						log.Printf("command failed: %+v\n", cse)
					}

					monitor.eventMu.Lock()
					monitor.commandFailed[cmd] = append(monitor.commandFailed[cmd], cse)
					monitor.eventMu.Unlock()
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

				monitor.eventMu.Lock()
				monitor.connectionCheckedIn[pe.ConnectionID] = append(monitor.connectionCheckedIn[pe.ConnectionID], pe)
				monitor.eventMu.Unlock()
			case event.ConnectionCheckedOut:
				if shouldLog {
					log.Printf("connection checked out: %+v\n", pe)
				}

				monitor.eventMu.Lock()
				monitor.connectionCheckedOut[pe.ConnectionID] = append(monitor.connectionCheckedOut[pe.ConnectionID], pe)
				monitor.eventMu.Unlock()
			case event.ConnectionClosed:
				if shouldLog {
					log.Printf("connection closed: %+v\n", pe)
				}

				monitor.eventMu.Lock()
				monitor.connectionClosed[pe.ConnectionID] = append(monitor.connectionClosed[pe.ConnectionID], pe)
				monitor.eventMu.Unlock()
			case event.ConnectionPendingReadStarted:
				if shouldLog {
					log.Printf("connection awaiting pending read: %+v\n", pe)
				}

				monitor.connectionPendingReadStarted[pe.ConnectionID] = append(monitor.connectionPendingReadStarted[pe.ConnectionID], pe)
				monitor.eventMu.Lock()
				monitor.eventMu.Unlock()
			case event.ConnectionPendingReadFailed:
				if shouldLog {
					log.Printf("connection pending read failed: %+v\n", pe)
				}

				monitor.eventMu.Lock()
				monitor.connectionPendingReadFailed[pe.ConnectionID] = append(monitor.connectionPendingReadFailed[pe.ConnectionID], pe)
				monitor.eventMu.Unlock()
			case event.ConnectionPendingReadSucceeded:
				if shouldLog {
					log.Printf("connection pending read succeeded: %+v\n", pe)
				}

				monitor.eventMu.Lock()
				monitor.connectionPendingReadSucceeded[pe.ConnectionID] = append(monitor.connectionPendingReadSucceeded[pe.ConnectionID], pe)
				monitor.eventMu.Unlock()
			}
		},
	}

	return monitor
}

func (m *monitor) Reset() {
	m.commandStarted = map[string][]*event.CommandStartedEvent{}
	m.commandFailed = map[string][]*event.CommandFailedEvent{}
	m.connectionClosed = map[int64][]*event.PoolEvent{}
	m.connectionCheckedIn = map[int64][]*event.PoolEvent{}
	m.connectionCheckedOut = map[int64][]*event.PoolEvent{}
	m.connectionPendingReadFailed = map[int64][]*event.PoolEvent{}
	m.connectionPendingReadStarted = map[int64][]*event.PoolEvent{}
	m.connectionPendingReadSucceeded = map[int64][]*event.PoolEvent{}
	m.eventMu = sync.Mutex{}
}
