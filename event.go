package eventsource

import (
	"bytes"
	"log"
)

type Event struct {
	Name     string
	Message  string
	Channels []string
	sent     int
	failed   int
}

type job struct {
	data []byte
	done chan bool
}

func (e *Event) loop(clients []*client) {
	pending := len(clients)
	if pending == 0 {
		return
	}
	done := make(chan bool, pending)
	p := job{data: e.bytes(), done: done}
	for i := range clients {
		c := clients[i]
		go func() {
			c.in <- p
		}()
	}
	for pending > 0 {
		ok := <-done
		if ok {
			e.sent++
		} else {
			e.failed++
		}
		pending--
	}
	log.Printf("{send: %d, failed: %d}\n", e.sent, e.failed)
}

func (e *Event) bytes() []byte {
	var buf bytes.Buffer
	if e.Name != "" {
		buf.WriteString("event: ")
		buf.WriteString(e.Name)
		buf.WriteString("\n")
	}
	buf.WriteString("data: ")
	buf.WriteString(e.Message)
	buf.WriteString("\n")
	return buf.Bytes()
}
