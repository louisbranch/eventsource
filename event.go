package eventsource

import (
	"bytes"
	"compress/zlib"
	"encoding/base64"
	"time"

	"log"
)

type event struct {
	name     string
	message  []byte
	channels []string
	compress bool
	started  time.Time
	finished time.Time
	sent     int
}

func newEvent(name string, message []byte, channels []string, compress bool) *event {
	e := event{
		name:     name,
		message:  message,
		channels: channels,
		compress: compress,
	}
	return &e
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
	if e.compress {
		deflated := deflate(e.message)
		buf.WriteString(deflated)
	} else {
		buf.Write(e.message)
	}
	buf.WriteString("\n\n")
	return buf.Bytes()
}

func deflate(message []byte) string {
	var buf bytes.Buffer
	w, _ := zlib.NewWriterLevel(&buf, 6)
	w.Write(message)
	w.Close()
	return base64.StdEncoding.EncodeToString(buf.Bytes())
}
