package eventsource

import (
	"bytes"
	"testing"
)

func TestEventBytesWithName(t *testing.T) {
	expecting := []byte("event: test\ndata: {id: 1}\n")
	e := event{
		name:    "test",
		message: "{id: 1}",
	}
	result := e.bytes()
	if !bytes.Equal(expecting, result) {
		t.Errorf("expected:\n%s\ngot:\n%s\n", expecting, result)
	}
}

func TestEventBytesWithoutName(t *testing.T) {
	expecting := []byte("data: {id: 1}\n")
	e := event{
		message: "{id: 1}",
	}
	result := e.bytes()
	if !bytes.Equal(expecting, result) {
		t.Errorf("expected:\n%s\ngot:\n%s\n", expecting, result)
	}
}
