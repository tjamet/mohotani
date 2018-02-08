package logger

// Logger is the interface that WatchListener loggers must implement.
// Typically, it can be instanciated as:
// import (
//		"log"
//		"os"
//	)
// log.New(os.Stdout, "IP Watcher:", log.LstdFlags)
type Logger interface {
	Printf(format string, v ...interface{})
}
