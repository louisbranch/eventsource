package eventsource

import (
	"bytes"
	"testing"
)

var message []byte = []byte("{id: 1}")

func TestEventBytesWithName(t *testing.T) {
	expecting := []byte("event: test\ndata: {id: 1}\n\n")
	e := event{
		name:    "test",
		message: message,
	}
	result := e.bytes()
	if !bytes.Equal(expecting, result) {
		t.Errorf("expected:\n%s\ngot:\n%s\n", expecting, result)
	}
}

func TestEventBytesWithoutName(t *testing.T) {
	expecting := []byte("data: {id: 1}\n\n")
	e := event{
		message: message,
	}
	result := e.bytes()
	if !bytes.Equal(expecting, result) {
		t.Errorf("expected:\n%s\ngot:\n%s\n", expecting, result)
	}
}

func TestEventBytesWithCompression(t *testing.T) {
	expecting := []byte("data: eJyqzkyxUjCsBQQAAP//CfUCUQ==\n\n")
	e := event{
		message:  message,
		compress: true,
	}
	result := e.bytes()
	if !bytes.Equal(expecting, result) {
		t.Errorf("expected:\n%s\ngot:\n%s\n", expecting, result)
	}
}
