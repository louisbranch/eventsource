package eventsource

import "time"

var ping = payload{data: []byte(":ping\n\n")}

// A server manages all clients, adding and removing them from the pool and
// receiving incoming events to forward to clients
type server struct {
	add    chan client
	remove chan client
	local  chan Event
	global chan Event
}

// The listen function is used to receive messages to add, remove and send
// events to clients. Every 30 seconds it sends a ping message to all
// clients to detect stale connections
func (s *server) listen() {
	var clients []client
	hearbeat := time.NewTicker(30 * time.Second)

	for {
		select {
		case c := <-s.add:
			clients = s.spawn(clients, c)
		case c := <-s.remove:
			clients = s.kill(clients, c)
		case e := <-s.local:
			s.send(clients, e)
		case e := <-s.global:
			s.broadcast(clients, e)
		case <-hearbeat.C:
			stats.ClientsCount(len(clients))
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

// The send function sends an event to all clients that have subscribed to one
// of the event's channels.
func (s *server) send(clients []client, e Event) {
	var subscribed []client
	for _, c := range clients {
		if hasChannel(c.channels, e.Channels) {
			subscribed = append(subscribed, c)
		}
	}
	if len(subscribed) > 0 {
		go e.send(subscribed)
	}
}

// The broadcast function sends an event to all clients.
func (s *server) broadcast(clients []client, e Event) {
	if len(clients) > 0 {
		go e.send(clients)
	}
}

// The ping functions writes to the stream of all clients to detect stale
// connections
func (s *server) ping(clients []client) {
	for _, c := range clients {
		go func(c client) {
			select {
			case c.events <- ping:
			case <-c.done:
			}
		}(c)
	}
}

// The hasChannel function returns whether client and server have a channel in
// common.
func hasChannel(client []string, server []string) bool {
	for _, c := range client {
		for _, s := range server {
			if c == s {
				return true
			}
		}
	}
	return false
}
