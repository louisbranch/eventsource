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
	"strings"
)

const (
	HEADER = `HTTP/1.1 200 OK
Content-Type: text/event-stream
Cache-Control: no-cache
Connection: keep-alive`

	BODY = "retry: 2000\n"
)

var padding string

type Server struct {
	maxClients int
	clients    []*client
}

//init generates a padding payload to establish a text/stream connection on
//Internet Explorer < 10. See
//http://blogs.msdn.com/b/ieinternals/archive/2010/04/06/comet-streaming-in-internet-explorer-with-xmlhttprequest-and-xdomainrequest.aspx
func init() {
	var buf bytes.Buffer
	buf.WriteByte(':')
	for i := 0; i < 2048; i++ {
		buf.WriteByte(' ')
	}
	buf.WriteByte('\n')
	padding = buf.String()
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

	channels := req.URL.Query().Get("channels")
	chans := strings.Split(channels, ",")
	client, err := s.add(conn, chans)
	if err != nil {
		conn.Write([]byte(err.Error()))
		conn.Close()
		log.Println(err)
		return
	}

	err = client.write(initialResponse(req))

	if err != nil {
		client.deactivate()
	}
}

// Broadcast sends a message to all active clients connected that subscribed to
// event's channel(s)
func (s *Server) Broadcast(e Event) {
	var subscribed []*client
	var inactives []*client

	for i := range s.clients {
		c := s.clients[i]
		if !c.active {
			inactives = append(inactives, c)
			continue
		}
		if contains(e.Channels, c.channels) {
			subscribed = append(subscribed, c)
		}
	}

	go e.loop(subscribed)

	for j := range inactives {
		c := inactives[j]
		s.remove(c)
		close(c.in)
	}
}

// add creates a new client for the connection and adds to the listening
// clients list, unless the max clients has been reached.
func (s *Server) add(conn net.Conn, channels []string) (*client, error) {
	l := len(s.clients)
	if l >= s.maxClients {
		return nil, errors.New("Max connections reached, closing connection.")
	}
	c := newClient(l, conn, channels)
	s.clients = append(s.clients, c)
	go c.loop()
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
// text/stream connection with retry option and CORS enabled.
func initialResponse(req *http.Request) []byte {
	var buf bytes.Buffer
	buf.WriteString(HEADER)
	if origin := req.Header.Get("origin"); origin != "" {
		cors := fmt.Sprintf("Access-Control-Allow-Origin: %s", origin)
		buf.WriteString("Access-Control-Allow-Credentials: true")
		buf.WriteString(cors)
	}
	buf.WriteString("\n\n")
	buf.WriteString(padding)
	buf.WriteString(BODY)
	return buf.Bytes()
}

// contains returns whether a string in a is also present in b
func contains(a []string, b []string) bool {
	match := false
loop:
	for i := range a {
		for j := range b {
			if a[i] == b[j] {
				match = true
				break loop
			}
		}
	}
	return match
}
