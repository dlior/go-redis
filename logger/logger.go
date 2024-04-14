package logger

import (
	"fmt"
	"time"
)

const (
	LogInfo    = "INFO"
	LogDebug   = "DEBUG"
	LogWarning = "WARNING"
	LogError   = "ERROR"
)

type LogEntry struct {
	Time     time.Time
	Severity string
	Message  string
}

var LogCh = make(chan LogEntry, 50)
var DoneCh = make(chan struct{})

func Logger() {

	for {
		select {
		case entry := <-LogCh:
			color := getLogColor(entry.Severity)
			fmt.Printf("%v\t\x1b[%dm[%v]: %v\x1b[0m\n", entry.Time.Format("2006-01-02 15:04:05"), color, entry.Severity, entry.Message)
		case <-DoneCh:
			return
		}
	}
}

func getLogColor(severity string) uint8 {
	switch severity {
	case LogError:
		return 31
	case LogInfo:
		return 32
	case LogWarning:
		return 33
	case LogDebug:
		return 34
	default:
		return 0
	}
}
