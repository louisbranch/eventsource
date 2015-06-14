package eventsource

import "testing"

func TestHasChannel(t *testing.T) {
	client1 := []string{"a", "b"}
	client2 := []string{"c", "d"}
	server := []string{"b", "e"}

	if !hasChannel(client1, server) {
		t.Errorf("expected:\n%q\nto be subscribed to:\n%q\n", client1, server)
	}

	if hasChannel(client2, server) {
		t.Errorf("expected:\n%q\nto not be subscribed to:\n%q\n", client2, server)
	}
}
