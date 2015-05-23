package eventsource

import "log"

type server struct {
	limit  int
	add    chan client
	remove chan client
	send   chan event
}

func (s *server) listen() {
	var clients []client
	for {
		select {
		case c := <-s.add:
			clients = s.spawn(clients, c)
		case c := <-s.remove:
			clients = s.kill(clients, c)
		case e := <-s.send:
			s.broadcast(clients, e)
		}
	}
}

func (s *server) spawn(clients []client, c client) []client {
	l := len(clients)
	go c.listen(s.remove)
	if l >= s.limit {
		c.events <- []byte(LIMIT_REACHED)
		close(c.events)
	} else {
		clients = append(clients, c)
	}
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
		log.Panic("Client not found")
	}

	if index < last {
		swap := clients[last]
		clients[index] = swap
	}
	clients = clients[:last]

	close(c.events)
	return clients
}

func (s *server) broadcast(clients []client, e event) {
	var subscribed []client
	for i := range clients {
		c := clients[i]
		if contains(c.channels, e.channels) {
			subscribed = append(subscribed, c)
		}
	}
	go e.send(subscribed)
}

// contains returns whether a string in a is also present in b
func contains(a []string, b []string) bool {
	for i := range a {
		for j := range b {
			if a[i] == b[j] {
				return true
			}
		}
	}
	return false
}
