package eventsource

import (
	"bytes"
	"testing"
)

func TestEventBytesWithName(t *testing.T) {
	expecting := []byte("event: test\ndata: {id: 1}\n")
	e := Event{
		Name:    "test",
		Message: "{id: 1}",
	}
	result := e.Bytes()
	if !bytes.Equal(expecting, result) {
		t.Errorf("expected:\n%s\ngot:\n%s\n", expecting, result)
	}
}

func TestEventBytesWithoutName(t *testing.T) {
	expecting := []byte("data: {id: 1}\n")
	e := Event{
		Message: "{id: 1}",
	}
	result := e.Bytes()
	if !bytes.Equal(expecting, result) {
		t.Errorf("expected:\n%s\ngot:\n%s\n", expecting, result)
	}
}
