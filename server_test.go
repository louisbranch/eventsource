package eventsource

import "testing"

func TestIsSubscribed(t *testing.T) {
	client1 := []string{"a", "b"}
	client2 := []string{"c", "d"}
	server := []string{"b", "e"}

	if !isSubscribed(client1, server) {
		t.Errorf("expected:\n%q\nto be subscribed to:\n%q\n", client1, server)
	}

	if isSubscribed(client2, server) {
		t.Errorf("expected:\n%q\nto be subscribed to:\n%q\n", client2, server)
	}
}
