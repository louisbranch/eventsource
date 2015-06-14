package eventsource

import "time"

// A server manages all clients, adding and removing them from the pool and
// receiving incoming events to forward to clients
type server struct {
	add    chan client
	remove chan client
	send   chan Event
}

// The listen function is used to receive messages to add, remove and broadcast
// events to client connected. Every 30 seconds it sends a ping message to all
// clients to detect stale connections
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
		}
	}
}

// The spawn function adds a new client to the clients list and launches a
// goroutine for the client to listen to incoming messages. The client receives
// the remove channel necessary to unsubscribe itself from the server.
func (s *server) spawn(clients []client, c client) []client {
	go c.listen(s.remove)
	clients = append(clients, c)
	return clients
}

// The kill function removes a client from the client list by comparing their
// events channel. The client is removed by being moved to the end of the list
// and reducing the slice length.
func (s *server) kill(clients []client, client client) []client {
	index := -1
	for i, c := range clients {
		if client.events == c.events {
			index = i
			break
		}
	}

	if index == -1 {
		panic("client not found")
	}

	last := len(clients) - 1
	if index < last {
		swap := clients[last]
		clients[index] = swap
	}
	clients = clients[:last]

	return clients
}

// The broadcast function sends an event to all clients connected that have
// subscribed to the same channels and the event being sent.
func (s *server) broadcast(clients []client, e Event) {
	var subscribed []client
	for _, c := range clients {
		if isSubscribed(c.channels, e.Channels) {
			subscribed = append(subscribed, c)
		}
	}
	if len(subscribed) > 0 {
		go e.send(subscribed)
	}
}

// The ping function sends a ping message to clients to detect stale connections
func (s *server) ping(clients []client) {
	if len(clients) > 0 {
		e := Event{}
		e.Message = []byte(PING)
		go e.send(clients)
	}
}

// The isSubscribed function returns whether a channel in a is also present in b
// or if both are empty
func isSubscribed(client []string, server []string) bool {
	if len(client) == 0 && len(server) == 0 {
		return true
	}
	for _, c := range client {
		for _, s := range server {
			if c == s {
				return true
			}
		}
	}
	return false
}
