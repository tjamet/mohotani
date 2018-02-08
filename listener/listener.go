package listener

import (
	"reflect"
	"time"
)

// Listener defines methods an object must implement to be used as a listener
type Listener interface {
	// Listen defines the method a Listener should implement to notify addition or removal
	// in the list of elements
	// Implementation should always post the whole new list of elements
	Listen(chan []string)
}

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

// Poll defines the interface a function should implement to be a poller
type Poll func() ([]string, error)

// PollListener a poller context
type PollListener struct {
	// Ticker is the channel controlling the polling interval (typically fed by time.NewTicker(1 * time.Second))
	Ticker <-chan time.Time
	// Logger is the logger in which errors and log messages will be printed
	Logger Logger
	// Poll is the function called to get the new state
	Poll Poll
}

// Listen implements the Listener interface and forwards all chandes to out
func (p *PollListener) Listen(out chan []string) {
	old, err := p.Poll()
	if err != nil {
		p.Logger.Printf("error: failed to resolve ip: %s", err.Error())
	} else {
		out <- old
	}
	for range p.Ticker {
		i, err := p.Poll()
		if err != nil {
			p.Logger.Printf("error: failed to resolve ip: %s", err.Error())
		} else {
			if !reflect.DeepEqual(i, old) {
				out <- i
				old = i
			}
		}
	}
}
