// Package eventsource provides a simple implementation for Server-sent events,
// an one-way stream to send data to browsers, see more at:
// (https://developer.mozilla.org/en-US/docs/Server-sent_events)
package eventsource

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
)

const (
	HEADER = `HTTP/1.1 200 OK
Content-Type: text/event-stream
Cache-Control: no-cache
Connection: keep-alive`

	BODY = "retry: 2000\n"
)

type Server struct {
	maxClients int
	clients    []*client
}

// New returns a new eventsource server with the maximum number of clients
// (connections) set.
func New(maxClients int) *Server {
	s := Server{maxClients: maxClients}
	return &s
}

// ServeHTTP implements the http handle interface.
// If the connection supports hijacking, it sends an initial header to switch
// to text/stream protocol and an initial body to retry after 2 seconds if the
// connection drops.
func (s *Server) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	hj, ok := res.(http.Hijacker)
	if !ok {
		http.Error(res, "webserver doesn't support hijacking", http.StatusInternalServerError)
		return
	}

	conn, _, err := hj.Hijack()
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}

	client, err := s.add(conn)
	if err != nil {
		log.Println(err)
		conn.Close()
		return
	}

	err = client.write(initialResponse(req))

	if err != nil {
		client.deactivate()
	}
}

// Broadcast sends a message to all active clients connected.
// TODO restrict the message to only the subscribed channels.
func (s *Server) Broadcast(name, message string, channels []string) {
	e := event{
		name:     name,
		message:  message,
		channels: channels,
	}

	inactives := []*client{}

	for i := range s.clients {
		c := s.clients[i]
		if c.active {
			select {
			case c.in <- e:
			default: //discard value
			}
		} else {
			inactives = append(inactives, c)
			close(c.in)
		}
	}

	for j := range inactives {
		s.remove(inactives[j])
	}
}

// add creates a new client for the connection and adds to the listening
// clients list, unless the max clients has been reached.
func (s *Server) add(conn net.Conn) (*client, error) {
	l := len(s.clients)
	if l >= s.maxClients {
		conn.Close()
		return nil, errors.New("Max connections reached, closing connection.")
	}
	c := newClient(l, conn)
	s.clients = append(s.clients, c)
	go c.listen()
	return c, nil
}

// remove removes a client from clients list.
// Note: the slice memory is not reclaimed
func (s *Server) remove(c *client) {
	l := len(s.clients) - 1
	i := c.index
	if i < l {
		swap := s.clients[l]
		s.clients[i] = swap
		swap.index = i
	}
	s.clients = s.clients[:l]
}

// initialResponse sends a header and body sent to client to establish a
// text/stream connection with retry option.
func initialResponse(req *http.Request) []byte {
	var buf bytes.Buffer
	buf.WriteString(HEADER)
	if origin := req.Header.Get("origin"); origin != "" {
		cors := fmt.Sprintf("Access-Control-Allow-Origin: %s", origin)
		buf.WriteString("Access-Control-Allow-Credentials: true")
		buf.WriteString(cors)
	}
	buf.WriteString("\n\n")
	buf.WriteString(BODY)
	return buf.Bytes()
}
