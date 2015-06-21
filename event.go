package eventsource

import (
	"bytes"
	"compress/zlib"
	"encoding/base64"
	"time"
)

// An event holds the data necessary to build the actual text/stream event
type Event struct {
	Name     string
	Message  []byte
	Channels []string
	Compress bool
}

// The send function receives a list of clients and send to them the text/stream
// event to be written on the client's connection. It returns the time spent to
// write to each client. 0 duration means an error.
func (e Event) send(clients []client) []time.Duration {
	durations := []time.Duration{}
	size := len(clients)
	done := make(chan time.Duration, size)
	p := payload{data: e.bytes(), done: done}

	for _, c := range clients {
		go func(c client) {
			select {
			case c.events <- p:
			case <-c.done:
				p.done <- 0
			}
		}(c)
	}

	for i := 0; i < size; i++ {
		s := <-done
		durations = append(durations, s)
	}

	return durations
}

// The bytes function returns the text/stream message to be sent to the client
// If the event has name, it is added first, then the data. Optionally, the data
// can be compressed using zlib
func (e Event) bytes() []byte {
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
