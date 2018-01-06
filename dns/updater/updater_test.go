package updater

import (
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type testUpdater struct {
	domains chan string
	ips     chan []string
	err     chan error
}

func (t *testUpdater) Update(domain string, ips ...string) error {
	t.domains <- domain
	t.ips <- ips
	return <-t.err
}

func (t *testUpdater) getDomains() []string {
	r := []string{}
	for {
		select {
		case d := <-t.domains:
			r = append(r, d)
		case <-time.After(10 * time.Millisecond):
			return r
		}
	}
}

func (t *testUpdater) getIPs() [][]string {
	r := [][]string{}
	for {
		select {
		case ip := <-t.ips:
			r = append(r, ip)
		case <-time.After(10 * time.Millisecond):
			return r
		}
	}
}

type testListener struct {
	c chan chan []string
}

func (l *testListener) Listen(c chan []string) {
	l.c <- c
}

func TestUpdater(t *testing.T) {
	failed := make(chan interface{})
	go func() {
		<-time.After(3 * time.Second)
		t.Error("timeout executing the test")
		for {
			failed <- nil
		}
	}()
	ipU := &testUpdater{
		domains: make(chan string, 10),
		ips:     make(chan []string, 10),
		err:     make(chan error, 10),
	}
	ipL := &testListener{
		c: make(chan chan []string),
	}
	dL := &testListener{
		c: make(chan chan []string),
	}
	u := Updater{
		Updater:        ipU,
		IPListener:     ipL,
		DomainListener: dL,
		Logger:         log.New(os.Stdout, "test logger: ", log.LstdFlags),
	}
	go u.Start()
	assert.Equal(t, []string{}, ipU.getDomains())
	assert.Equal(t, [][]string{}, ipU.getIPs())

	ipListenerChannel := <-ipL.c
	domainListenerChannel := <-dL.c

	ipU.err <- nil
	select {
	case ipListenerChannel <- []string{"127.0.0.1"}:
	case <-failed:
		t.Errorf("Failed to notify IP listeners")
	}
	assert.Equal(t, []string{}, ipU.getDomains())
	assert.Equal(t, [][]string{}, ipU.getIPs())

	ipU.err <- nil
	select {
	case ipListenerChannel <- []string{"169.0.0.1"}:
	case <-failed:
		t.Errorf("Failed to notify IP listeners")
	}
	assert.Equal(t, []string{}, ipU.getDomains())
	assert.Equal(t, [][]string{}, ipU.getIPs())

	ipU.err <- nil
	select {
	case domainListenerChannel <- []string{"www.example.com"}:
	case <-failed:
		t.Errorf("Failed to notify Domain listeners")
	}
	assert.Equal(t, []string{"www.example.com"}, ipU.getDomains())
	assert.Equal(t, [][]string{{"169.0.0.1"}}, ipU.getIPs())

	ipU.err <- nil
	select {
	case ipListenerChannel <- []string{"127.0.0.1", "10.2.0.1"}:
	case <-failed:
		t.Errorf("Failed to notify IP listeners")
	}
	assert.Equal(t, []string{"www.example.com"}, ipU.getDomains())
	assert.Equal(t, [][]string{{"127.0.0.1", "10.2.0.1"}}, ipU.getIPs())

	ipU.err <- fmt.Errorf("test error")
	select {
	case ipListenerChannel <- []string{"127.0.0.1", "10.2.0.1"}:
	case <-failed:
		t.Errorf("Failed to notify IP listeners")
	}
	assert.Equal(t, []string{"www.example.com"}, ipU.getDomains())
	assert.Equal(t, [][]string{{"127.0.0.1", "10.2.0.1"}}, ipU.getIPs())

	// can recover from errors
	ipU.err <- nil
	select {
	case domainListenerChannel <- []string{"www.example.com", "www2.example.com"}:
	case <-failed:
		t.Errorf("Failed to notify Domain listeners")
	}
	assert.Equal(t, []string{"www.example.com", "www2.example.com"}, ipU.getDomains())
	assert.Equal(t, [][]string{{"127.0.0.1", "10.2.0.1"}, {"127.0.0.1", "10.2.0.1"}}, ipU.getIPs())
}
