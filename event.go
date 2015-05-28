package eventsource

import (
	"bytes"
	"time"

	"log"
)

type event struct {
	name     string
	message  string
	channels []string
	started  time.Time
	finished time.Time
	sent     int
}

func (e *event) send(clients []client) {
	e.started = time.Now()
	defer func() {
		e.finished = time.Now()
		duration := time.Since(e.started)
		log.Printf("[INFO] event sent=%d duration=%d", e.sent, duration)
	}()

	pending := len(clients)

	if pending == 0 {
		return
	}

	done := make(chan bool, pending)
	payload := e.bytes()

	for i := range clients {
		c := clients[i]
		go func() {
			select {
			case c.events <- payload:
				done <- true
			case <-c.done:
				done <- false
			}
		}()
	}

	for pending > 0 {
		ok := <-done
		if ok {
			e.sent++
		}
		pending--
	}
}

func (e *event) bytes() []byte {
	var buf bytes.Buffer
	if e.name != "" {
		buf.WriteString("event: ")
		buf.WriteString(e.name)
		buf.WriteString("\n")
	}
	buf.WriteString("data: ")
	buf.WriteString(e.message)
	buf.WriteString("\n")
	return buf.Bytes()
}
