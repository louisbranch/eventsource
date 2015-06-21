package eventsource

import (
	"bytes"
	"net"
	"reflect"
	"testing"
)

var message []byte = []byte("{id: 1}")
var deflated = "eJyqzkyxUjCsBQQAAP//CfUCUQ=="

/*
func TestDefaultEventSendToClient(t *testing.T) {
	e := DefaultEvent{Message: message}
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

func TestDefaultEventSendDuration(t *testing.T) {
	e := DefaultEvent{Message: message}
	c := stubClient()
	go c.listen(make(chan client))

	result := e.send([]client{c})
	if len(result) != 1 {
		t.Errorf("expected: 1 duration%s\ngot:\n%s\n", result)
	}
}

func TestDefaultEventSendError(t *testing.T) {
	e := DefaultEvent{Message: message}
	c := stubClient()
	close(c.done)
	result := e.send([]client{c})

	expecting := []time.Duration{0}
	if !reflect.DeepEqual(expecting, result) {
		t.Errorf("expected:\n%s\ngot:\n%s\n", expecting, result)
	}
}

*/

func TestDefaultEventBytesWithId(t *testing.T) {
	expecting := []byte("id: 1\ndata: {id: 1}\n\n")
	e := DefaultEvent{
		Id:      1,
		Message: message,
	}
	result := e.Bytes()
	if !bytes.Equal(expecting, result) {
		t.Errorf("expected:\n%s\ngot:\n%s\n", expecting, result)
	}
}

func TestDefaultEventBytesWithName(t *testing.T) {
	expecting := []byte("event: test\ndata: {id: 1}\n\n")
	e := DefaultEvent{
		Name:    "test",
		Message: message,
	}
	result := e.Bytes()
	if !bytes.Equal(expecting, result) {
		t.Errorf("expected:\n%s\ngot:\n%s\n", expecting, result)
	}
}

func TestDefaultEventBytesWithoutName(t *testing.T) {
	expecting := []byte("data: {id: 1}\n\n")
	e := DefaultEvent{
		Message: message,
	}
	result := e.Bytes()
	if !bytes.Equal(expecting, result) {
		t.Errorf("expected:\n%s\ngot:\n%s\n", expecting, result)
	}
}

func TestDefaultEventBytesWithCompression(t *testing.T) {
	expecting := []byte("data: " + deflated + "\n\n")
	e := DefaultEvent{
		Message:  message,
		Compress: true,
	}
	result := e.Bytes()
	if !bytes.Equal(expecting, result) {
		t.Errorf("expected:\n%s\ngot:\n%s\n", expecting, result)
	}
}

func TestDefaultEventDeflate(t *testing.T) {
	expecting := deflated
	e := DefaultEvent{Message: message}
	result := e.deflate()
	if expecting != result {
		t.Errorf("expected:\n%s\ngot:\n%s\n", expecting, result)
	}
}

func TestDefaultEventClientsWithNoChannel(t *testing.T) {
	client1 := client{channels: []string{"a", "b"}}
	client2 := client{channels: []string{"c", "d"}}
	e := DefaultEvent{}

	expected := []client{client1, client2}
	result := e.Clients([]client{client1, client2})

	if !reflect.DeepEqual(expected, result) {
		t.Errorf("expected:\n%v\nto be equal to:\n%v\n", expected, result)
	}
}

func TestDefaultEventClientsWithChannels(t *testing.T) {
	client1 := client{channels: []string{"a", "b"}}
	client2 := client{channels: []string{"c", "d"}}
	e := DefaultEvent{Channels: []string{"b", "e"}}

	expected := []client{client1}
	result := e.Clients([]client{client1, client2})

	if !reflect.DeepEqual(expected, result) {
		t.Errorf("expected:\n%v\nto be equal to:\n%v\n", expected, result)
	}
}

func TestPingBytes(t *testing.T) {
	expecting := []byte(":ping\n\n")
	result := ping{}.Bytes()
	if !bytes.Equal(expecting, result) {
		t.Errorf("expected:\n%s\ngot:\n%s\n", expecting, result)
	}
}

func TestPingClients(t *testing.T) {
	clients := []client{client{}}
	expecting := clients
	result := ping{}.Clients(clients)
	if !reflect.DeepEqual(expecting, result) {
		t.Errorf("expected:\n%v\ngot:\n%v\n", expecting, result)
	}
}

func stubClient() client {
	net.Listen("tcp4", "127.0.0.1:4000")
	conn, _ := net.Dial("tcp4", "127.0.0.1:4000")
	c := client{events: make(chan payload), conn: conn, done: make(chan bool)}
	return c
}
