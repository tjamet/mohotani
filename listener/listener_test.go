package listener

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type testPoller struct {
	IPs <-chan []string
	err <-chan error
}

func (r *testPoller) Poll() ([]string, error) {
	return <-r.IPs, <-r.err
}

type testLogger struct {
	messages chan string
}

func (l *testLogger) Printf(format string, v ...interface{}) {
	l.messages <- fmt.Sprintf(format, v...)
}

func TestWatch(t *testing.T) {
	l := &testLogger{
		make(chan string),
	}
	ticker := make(chan time.Time)
	out := make(chan []string)

	pollIPs := make(chan []string, 10)
	pollErrs := make(chan error, 10)
	r := &testPoller{
		IPs: pollIPs,
		err: pollErrs,
	}

	listener := &PollListener{
		Logger: l,
		Ticker: ticker,
		Poll:   r.Poll,
	}
	go listener.Listen(out)

	pollIPs <- []string{"value 1"}
	pollErrs <- nil

	ticker <- time.Now()
	select {
	case out := <-out:
		assert.Equal(t, []string{"value 1"}, out)
	case m := <-l.messages:
		t.Errorf("Unexpected message in listener %s", m)
	case <-time.After(3 * time.Second):
		t.Error("Timeout reading the output channel")
	}

	pollIPs <- []string{"value 2"}
	pollErrs <- nil

	ticker <- time.Now()
	select {
	case out := <-out:
		assert.Equal(t, []string{"value 2"}, out)
	case m := <-l.messages:
		t.Errorf("Unexpected message in listener logger %s", m)
	case <-time.After(3 * time.Second):
		t.Error("Timeout reading the output channel")
	}

	pollIPs <- []string{"value 2"}
	pollErrs <- nil
	select {
	case out := <-out:
		t.Errorf("An ip value %s value was output without a timer tick", out)
	case m := <-l.messages:
		t.Errorf("Unexpected message in listener logger %s", m)
	case <-time.After(100 * time.Millisecond):
	}

	pollIPs <- []string{"value 2"}
	pollErrs <- nil
	ticker <- time.Now()
	select {
	case out := <-out:
		t.Errorf("An ip value %s value was output with no changes in IPs", out)
	case m := <-l.messages:
		t.Errorf("Unexpected message in listener logger %s", m)
	case <-time.After(100 * time.Millisecond):
	}

	pollIPs <- []string{"This should never be forwarded"}
	pollErrs <- fmt.Errorf("test eror")
	ticker <- time.Now()
	select {
	case out := <-out:
		t.Errorf("An ip value %s value was output with a resolver in error", out)
	case m := <-l.messages:
		assert.Equal(t, "error: failed to resolve ip: test eror", m)
	case <-time.After(100 * time.Millisecond):
	}
}
