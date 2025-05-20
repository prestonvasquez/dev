package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

type ProgressName struct {
	Name     string
	Progress float64
}

func main() {
	logrus.SetOutput(os.Stdout)
	logrus.SetLevel(logrus.InfoLevel)

	files := []string{"foo.txt", "bar.log", "baz.csv", "qux.bin"}
	pn := make(chan ProgressName)
	// Start the logger listener
	go pullWithProgress(len(files), pn)

	for _, file := range files {
		for i := 0; i <= 100; i++ {
			time.Sleep(10 * time.Millisecond)

			pn <- ProgressName{
				Name:     file,
				Progress: float64(i),
			}
		}
	}

	close(pn)
}

// pullWithProgress reads from a channel of ProgressName and updates a single line in-place
func pullWithProgress(n int, progressCh <-chan ProgressName) {
	formatter, ok := logrus.StandardLogger().Formatter.(*logrus.TextFormatter)
	if !ok {
		// fallback: simple logging
		for pr := range progressCh {
			logrus.Infof("%s: %6.2f%%", pr.Name, pr.Progress)
		}
		return
	}

	logrus.Infof("📥 Pulling data") // Update this line to update percentage over the entire progressCh

	var oldName string
	count := 1

	// Loop over progress events
	for pr := range progressCh {
		if oldName != "" && oldName != pr.Name {
			// Break for each new file.
			os.Stdout.Write([]byte("\n"))
			count++
		}
		oldName = pr.Name

		// Build log entry
		entry := &logrus.Entry{
			Logger:  logrus.StandardLogger(),
			Data:    logrus.Fields{},
			Time:    time.Now(),
			Level:   logrus.InfoLevel,
			Message: fmt.Sprintf("  [%d/%d] %s: %6.2f%%", count, n, pr.Name, pr.Progress),
		}

		// Format without newline
		lineBytes, err := formatter.Format(entry)
		if err != nil {
			continue
		}
		line := strings.TrimRight(string(lineBytes), "\n")

		// Carriage return + overwrite
		os.Stdout.Write([]byte("\r" + line))
	}

	// After channel closes, finalize with newline
	os.Stdout.Write([]byte("\n"))
}
