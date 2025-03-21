package metrics

import (
	"context"
	"sync/atomic"

	"go.mongodb.org/mongo-driver/event"
)

type CommandMonitor struct {
	*event.CommandMonitor

	Failed    atomic.Int32
	Succeeded atomic.Int32
	Started   atomic.Int32
}

type CommandMonitorCallback func() *CommandMonitor

var _ CommandMonitorCallback = newExpFuncCommandMonitor

func newExpFuncCommandMonitor() *CommandMonitor {
	commandMonitor := &CommandMonitor{}

	commandMonitor.CommandMonitor = &event.CommandMonitor{
		Started: func(_ context.Context, cse *event.CommandStartedEvent) {
			if cse.CommandName == "find" {
				commandMonitor.Started.Add(1)
			}
		},
		Succeeded: func(_ context.Context, cse *event.CommandSucceededEvent) {
			if cse.CommandName == "find" {
				commandMonitor.Succeeded.Add(1)
			}
		},
		Failed: func(_ context.Context, evt *event.CommandFailedEvent) {
			if evt.CommandName == "find" {
				commandMonitor.Failed.Add(1)
			}
		},
	}

	return commandMonitor
}
