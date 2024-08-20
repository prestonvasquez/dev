package v2

import (
	"context"
	"log"

	"go.mongodb.org/mongo-driver/v2/event"
)

// commandMonitorByName will create a command monitor that logs commands for
// the specific list of names.
func commandMonitorByName(log *log.Logger, cmdNames ...string) *event.CommandMonitor {
	monitor := &event.CommandMonitor{
		Started: func(_ context.Context, cse *event.CommandStartedEvent) {
			for _, cmdName := range cmdNames {
				if cse.CommandName == cmdName {
					log.Printf("command started: %+v\n", cse.Command)
				}
			}
		},
		Succeeded: func(_ context.Context, cse *event.CommandSucceededEvent) {
			for _, cmdName := range cmdNames {
				if cse.CommandName == cmdName {
					log.Printf("command succeeded: %+v\n", cse.Reply)
				}
			}
		},
		Failed: func(_ context.Context, cfe *event.CommandFailedEvent) {
			for _, cmdName := range cmdNames {
				if cfe.CommandName == cmdName {
					log.Printf("command failed: %+v\n", cfe.Failure)
				}
			}
		},
	}

	return monitor
}
