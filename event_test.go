package eventsource

import (
	"bytes"
	"net"
	"reflect"
	"testing"
	"time"
)

var message []byte = []byte("{id: 1}")
var deflated = "eJyqzkyxUjCsBQQAAP//CfUCUQ=="

func TestEventSendToClient(t *testing.T) {
	e := Event{Message: message}
	c := client{events: make(chan payload)}
	go c.listen(make(chan client))
	go e.send([]client{c})
	p := <-c.events

	expecting := e.bytes()
	result := p.data

	if !bytes.Equal(expecting, result) {
		t.Errorf("expected:\n%s\ngot:\n%s\n", expecting, result)
	}
}

func TestEventSendDuration(t *testing.T) {
	e := Event{Message: message}
	c := stubClient()
	go c.listen(make(chan client))

	result := e.send([]client{c})
	if len(result) != 1 {
		t.Errorf("expected: 1 duration%s\ngot:\n%s\n", result)
	}
}

func TestEventSendError(t *testing.T) {
	e := Event{Message: message}
	c := stubClient()
	close(c.done)
	result := e.send([]client{c})

	expecting := []time.Duration{0}
	if !reflect.DeepEqual(expecting, result) {
		t.Errorf("expected:\n%s\ngot:\n%s\n", expecting, result)
	}
}

func TestEventBytesWithName(t *testing.T) {
	expecting := []byte("event: test\ndata: {id: 1}\n\n")
	e := Event{
		Name:    "test",
		Message: message,
	}
	result := e.bytes()
	if !bytes.Equal(expecting, result) {
		t.Errorf("expected:\n%s\ngot:\n%s\n", expecting, result)
	}
}

func TestEventBytesWithoutName(t *testing.T) {
	expecting := []byte("data: {id: 1}\n\n")
	e := Event{
		Message: message,
	}
	result := e.bytes()
	if !bytes.Equal(expecting, result) {
		t.Errorf("expected:\n%s\ngot:\n%s\n", expecting, result)
	}
}

func TestEventBytesWithCompression(t *testing.T) {
	expecting := []byte("data: " + deflated + "\n\n")
	e := Event{
		Message:  message,
		Compress: true,
	}
	result := e.bytes()
	if !bytes.Equal(expecting, result) {
		t.Errorf("expected:\n%s\ngot:\n%s\n", expecting, result)
	}
}

func TestDeflate(t *testing.T) {
	expecting := deflated
	result := deflate(message)
	if expecting != result {
		t.Errorf("expected:\n%s\ngot:\n%s\n", expecting, result)
	}
}

func stubClient() client {
	net.Listen("tcp4", "127.0.0.1:4000")
	conn, _ := net.Dial("tcp4", "127.0.0.1:4000")
	c := client{events: make(chan payload), conn: conn, done: make(chan bool)}
	return c
}
