package eventsource

import (
	"bytes"
	"compress/zlib"
	"encoding/base64"
	"time"

	"log"
)

// An event is the high-level construct to send messages to clients
// It holds all the information necessary to build the actual text/stream event
type Event struct {
	Name     string
	Message  []byte
	Channels []string
	Compress bool
}

// The send function receives a list of clients and send to each client channel
// the text/stream event to be written on the client's connection
// After all clients have received the message or disconnected, the event logs
// how many clients received the message and the duration the event took to
// finish
func (e *Event) send(clients []client) {
	pending := len(clients)
	if pending == 0 {
		return
	}

	started := time.Now()

	done := make(chan bool, pending)
	data := e.bytes()

	for i := range clients {
		c := clients[i]
		go func() {
			select {
			case c.events <- data:
				done <- true
			case <-c.done:
				done <- false
			}
		}()
	}

	sent := 0
	for pending > 0 {
		ok := <-done
		if ok {
			sent++
		}
		pending--
	}

	duration := time.Since(started)
	log.Printf("[INFO] event sent=%d duration=%d", sent, duration)
}

// The bytes function returns the text/stream message to be sent to the client
// If the event has name, it is added first, then the data. Optionally, the data
// can be compressed using zlib
func (e *Event) bytes() []byte {
	var buf bytes.Buffer
	if e.Name != "" {
		buf.WriteString("event: ")
		buf.WriteString(e.Name)
		buf.WriteString("\n")
	}
	buf.WriteString("data: ")
	if e.Compress {
		deflated := deflate(e.Message)
		buf.WriteString(deflated)
	} else {
		buf.Write(e.Message)
	}
	buf.WriteString("\n\n")
	return buf.Bytes()
}

// The deflate function compress a slice of bytes using zlib default compression
// and returns a base64 encoded string
func deflate(message []byte) string {
	var buf bytes.Buffer
	w := zlib.NewWriter(&buf)
	w.Write(message)
	w.Close()
	return base64.StdEncoding.EncodeToString(buf.Bytes())
}
