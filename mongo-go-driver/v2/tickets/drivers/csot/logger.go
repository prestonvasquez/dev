package csot

import (
	"log"
	"sync"
	"testing"
)

const (
	commandFailed                  = "Command failed"
	commandStarted                 = "Command started"
	commandSucceeded               = "Command succeeded"
	connectionCheckedOut           = "Connection checked out"
	connectionCheckedIn            = "Connection checked in"
	connectionPendingReadStarted   = "Pending read started"
	connectionPendingReadSucceeded = "Pending read succeeded"
	connectionPendingReadFailed    = "Pending read failed"
	connectionClosed               = "Connection closed"
)

type csotLogger struct {
	mu sync.Mutex
	t  *testing.T

	logIO bool
	on    bool

	commandStartedCount                 int
	commandFailedCount                  int
	commandSucceededCount               int
	connectionCheckedOutCount           int
	connectionCheckedInCount            int
	connectionPendingReadStartedCount   int
	connectionPendingReadSucceededCount int
	connectionPendingReadFailedCount    int
	connectionClosedCount               int
}

func newCustomLogger(t *testing.T, logIO bool) *csotLogger {
	l := &csotLogger{logIO: logIO, t: t}

	return l
}

func (logger *csotLogger) toggleOn() {
	logger.on = true
}

func (logger *csotLogger) toggleOff() {
	logger.on = false
}

func (logger *csotLogger) Info(level int, msg string, kv ...interface{}) {
	logger.mu.Lock()
	defer logger.mu.Unlock()

	if !logger.on {
		return
	}

	printLog := func() {
		if logger.logIO {
			logger.t.Logf("level=%d msg=%s kv=%v\n", level, msg, kv)
		}
	}

	switch msg {
	case commandFailed:
		logger.commandFailedCount++

		printLog()
	case commandStarted:
		logger.commandStartedCount++

		printLog()
	case commandSucceeded:
		logger.commandSucceededCount++

		printLog()
	case connectionCheckedIn:
		logger.connectionCheckedInCount++

		printLog()
	case connectionCheckedOut:
		logger.connectionCheckedOutCount++

		printLog()
	case connectionClosed:
		logger.connectionClosedCount++

		printLog()
	case connectionPendingReadFailed:
		logger.connectionPendingReadFailedCount++

		printLog()
	case connectionPendingReadStarted:
		logger.connectionPendingReadStartedCount++

		printLog()
	case connectionPendingReadSucceeded:
		logger.connectionPendingReadSucceededCount++

		printLog()
	}
}

func (logger *csotLogger) Error(err error, msg string, _ ...interface{}) {
	logger.mu.Lock()
	defer logger.mu.Unlock()

	log.Printf("err=%v msg=%s\n", err, msg)
}
