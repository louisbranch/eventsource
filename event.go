package eventsource

import (
	"bytes"
	"compress/zlib"
	"encoding/base64"
	"strconv"
)

type Event interface {
	// The Bytes function returns the data to be written on the clients connection
	Bytes() []byte

	// The Clients function receives a list of clients and return a filtered list
	// of clients.
	Clients([]client) []client
}

// An event holds the data necessary to build the actual text/stream event
type DefaultEvent struct {
	Id       int
	Name     string
	Message  []byte
	Channels []string
	Compress bool
}

// The bytes function returns the text/stream message to be sent to the client.
// If the event has name, it is added first, then the data. Optionally, the data
// can be compressed using zlib.
func (e DefaultEvent) Bytes() []byte {
	var buf bytes.Buffer
	if e.Id > 0 {
		buf.WriteString("id: ")
		buf.WriteString(strconv.Itoa(e.Id))
		buf.WriteString("\n")
	}
	if e.Name != "" {
		buf.WriteString("event: ")
		buf.WriteString(e.Name)
		buf.WriteString("\n")
	}
	buf.WriteString("data: ")
	if e.Compress {
		buf.WriteString(e.deflate())
	} else {
		buf.Write(e.Message)
	}
	buf.WriteString("\n\n")
	return buf.Bytes()
}

// The Clients function selects clients that have at least one channel in
// common with the event or all clients if the event has no channel.
func (e DefaultEvent) Clients(clients []client) []client {
	if len(e.Channels) == 0 {
		return clients
	}
	var subscribed []client
	for _, client := range clients {
	channels:
		for _, cChans := range client.channels {
			for _, eChans := range e.Channels {
				if cChans == eChans {
					subscribed = append(subscribed, client)
					break channels
				}
			}
		}
	}
	return subscribed
}

// The deflate function compress the event message using zlib default
// compression and returns a base64 encoded string.
func (e DefaultEvent) deflate() string {
	var buf bytes.Buffer
	w := zlib.NewWriter(&buf)
	w.Write(e.Message)
	w.Close()
	return base64.StdEncoding.EncodeToString(buf.Bytes())
}

type ping struct{}

func (ping) Bytes() []byte {
	return []byte(":ping\n\n")
}

func (ping) Clients(clients []client) []client {
	return clients
}
