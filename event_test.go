package eventsource

import (
	"bytes"
	"testing"
)

var message []byte = []byte("{id: 1}")
var deflated = "eJyqzkyxUjCsBQQAAP//CfUCUQ=="

func TestEventSendToClient(t *testing.T) {
	e := Event{Message: message}
	c := client{events: make(chan payload)}
	go e.send([]client{c})
	p := <-c.events

	expecting := e.bytes()
	result := p.data

	if !bytes.Equal(expecting, result) {
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
