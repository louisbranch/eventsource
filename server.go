package eventsource

import (
	"time"

	"log"
)

// A server manages all clients, adding and removing them from the pool and
// receiving incoming events to forward to clients
type server struct {
	add    chan client
	remove chan client
	send   chan Event
}

func (s *server) listen() {
	hearbeat := time.NewTicker(30 * time.Second)
	var clients []client
	for {
		select {
		case c := <-s.add:
			clients = s.spawn(clients, c)
		case c := <-s.remove:
			clients = s.kill(clients, c)
		case e := <-s.send:
			s.broadcast(clients, e)
		case <-hearbeat.C:
			s.ping(clients)
			log.Printf("[INFO] clients count=%d", len(clients))
		}
	}
}

func (s *server) spawn(clients []client, c client) []client {
	go c.listen(s.remove)
	clients = append(clients, c)
	return clients
}

func (s *server) kill(clients []client, c client) []client {
	last := len(clients) - 1
	index := -1

	for i := range clients {
		if c.events == clients[i].events {
			index = i
			break
		}
	}

	if index == -1 {
		log.Println("[ERROR] client not found to be removed")
		return clients
	}

	if index < last {
		swap := clients[last]
		clients[index] = swap
	}
	clients = clients[:last]

	return clients
}

func (s *server) broadcast(clients []client, e Event) {
	var subscribed []client
	for i := range clients {
		c := clients[i]
		if isSubscribed(c.channels, e.Channels) {
			subscribed = append(subscribed, c)
		}
	}
	go e.send(subscribed)
}

func (s *server) ping(clients []client) {
	msg := []byte(PING)
	for i := range clients {
		c := clients[i]
		go func() {
			select {
			case c.events <- msg:
			case <-c.done:
			}
		}()
	}
}

// The isSubscribed function returns whether a channel in a is also present in b
func isSubscribed(a []string, b []string) bool {
	for i := range a {
		for j := range b {
			if a[i] == b[j] {
				return true
			}
		}
	}
	return false
}
