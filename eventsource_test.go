package eventsource

import (
	"reflect"
	"testing"
	"time"
)

func TestEventsourceStartNoChannels(t *testing.T) {
	es := Eventsource{}
	es.Start()
	result, ok := es.ChanSub.(NoChannels)
	if !ok {
		t.Errorf("expected to be NoChannels\ngot:\n%T\n", result)
	}
}

func TestEventsourceStartDefaultHttpOptions(t *testing.T) {
	es := Eventsource{}
	es.Start()
	opts, ok := es.HttpOptions.(DefaultHttpOptions)
	if !ok {
		t.Errorf("expected to be DefaultHttpOptions\ngot:\n%T\n", opts)
	}
	expecting := 2000
	retry := opts.Retry
	if expecting != retry {
		t.Errorf("expected retry to be:\n%d\ngot:\n%d\n", expecting, retry)
	}
	cors := opts.Cors
	if !cors {
		t.Errorf("expected Cors to be:\ntrue\ngot:\n%d\n", cors)
	}
	old := opts.OldBrowserSupport
	if !old {
		t.Errorf("expected OldBrowserSupport to be:\n%true\ngot:\n%d\n", old)
	}
}

func TestEventsourceStartServerChannels(t *testing.T) {
	es := Eventsource{}
	es.Start()
	s := es.server

	if s.add == nil {
		t.Errorf("expected server add channel to be created")
	}
	if s.remove == nil {
		t.Errorf("expected server remove channel to be created")
	}
	if s.events == nil {
		t.Errorf("expected server events channel to be created")
	}
}

func TestEventsourceStartServerListen(t *testing.T) {
	es := Eventsource{}
	es.Start()
	timeout := make(chan bool, 1)
	go func() {
		time.Sleep(1 * time.Second)
		timeout <- true
	}()
	select {
	case es.server.events <- DefaultEvent{}:
	case <-timeout:
		t.Errorf("expected server to be listening to events")
	}
}

func TestEventsourceSend(t *testing.T) {
	es := Eventsource{}
	events := make(chan Event, 1)
	es.server = server{events: events}
	expecting := DefaultEvent{Name: "test"}
	es.Send(expecting)
	result := <-events
	if !reflect.DeepEqual(expecting, result) {
		t.Errorf("expected:\n%v\nto be equal to:\n%v\n", expecting, result)
	}
}
