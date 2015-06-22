package eventsource

import (
	"bytes"
	"net"
	"reflect"
	"testing"
	"time"
)

func TestAddChannel(t *testing.T) {
	s := server{add: make(chan client)}
	c := client{}
	go s.listen()
	timeout := make(chan bool, 1)
	go func() {
		time.Sleep(1 * time.Second)
		timeout <- true
	}()
	select {
	case s.add <- c:
	case <-timeout:
		t.Errorf("expected server to be listening to add channel")
	}
}

func TestRemoveChannel(t *testing.T) {
	s := server{add: make(chan client), remove: make(chan client)}
	c := client{}
	go s.listen()
	s.add <- c
	timeout := make(chan bool, 1)
	go func() {
		time.Sleep(1 * time.Second)
		timeout <- true
	}()
	select {
	case s.remove <- c:
	case <-timeout:
		t.Errorf("expected server to be listening to remove channel")
	}
}

func TestServerPing(t *testing.T) {
	s := server{hearbeat: 1 * time.Nanosecond, add: make(chan client)}
	e := ping{}
	c := client{events: make(chan payload)}
	go s.listen()
	s.add <- c
	p := <-c.events

	expecting := e.Bytes()
	result := p.data

	if !bytes.Equal(expecting, result) {
		t.Errorf("expected:\n%s\ngot:\n%s\n", expecting, result)
	}
}

func TestServerSpawn(t *testing.T) {
	s := server{}
	c := client{}
	expecting := []client{c}
	result := s.spawn([]client{}, c)
	if !reflect.DeepEqual(expecting, result) {
		t.Errorf("expected:\n%v\nto be equal to:\n%v\n", expecting, result)
	}
}

func TestServerKill(t *testing.T) {
	s := server{}
	c1 := client{events: make(chan payload)}
	c2 := client{events: make(chan payload)}
	expecting := []client{c2}
	result := s.kill([]client{c1, c2}, c1)
	if !reflect.DeepEqual(expecting, result) {
		t.Errorf("expected:\n%v\nto be equal to:\n%v\n", expecting, result)
	}
}

func TestServerKillPanic(t *testing.T) {
	c1 := client{events: make(chan payload)}
	c2 := client{events: make(chan payload)}
	s := server{}
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("expected function to panic, it did not\n")
		}
	}()
	s.kill([]client{c2}, c1)
}

func TestServerSendPayload(t *testing.T) {
	s := server{}
	e := DefaultEvent{Message: message}
	c := client{events: make(chan payload)}
	go c.listen(make(chan client))
	go s.send(e, []client{c})
	p := <-c.events

	expecting := e.Bytes()
	result := p.data

	if !bytes.Equal(expecting, result) {
		t.Errorf("expected:\n%s\ngot:\n%s\n", expecting, result)
	}
}

func TestServerSendDuration(t *testing.T) {
	s := server{}
	e := DefaultEvent{Message: message}
	c := stubTCPClient()
	go c.listen(make(chan client))

	result := s.send(e, []client{c})
	if len(result) != 1 {
		t.Errorf("expected: 1 duration%s\ngot:\n%s\n", result)
	}
}

func TestServerSendError(t *testing.T) {
	s := server{}
	e := DefaultEvent{Message: message}
	c := stubTCPClient()
	close(c.done)
	result := s.send(e, []client{c})

	expecting := []time.Duration{0}
	if !reflect.DeepEqual(expecting, result) {
		t.Errorf("expected:\n%s\ngot:\n%s\n", expecting, result)
	}
}

func stubTCPClient() client {
	net.Listen("tcp4", "127.0.0.1:4000")
	conn, _ := net.Dial("tcp4", "127.0.0.1:4000")
	c := client{events: make(chan payload), conn: conn, done: make(chan bool)}
	return c
}
