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
Connection: keep-alive
Access-Control-Allow-Credentials: true`

	BODY = "\n\nretry: 2000\n"
)

type server struct {
	maxClients int
	clients    []*client
}

func New(maxClients int) *server {
	s := server{maxClients: maxClients}
	return &s
}

func (s *server) ServeHTTP(res http.ResponseWriter, req *http.Request) {
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

func (s *server) Broadcast(message, name string) {
	e := event{
		name:    name,
		message: []byte(message + "\n"),
	}

	inactives := []*client{}

	for i := range s.clients {
		c := s.clients[i]
		if c.active {
			c.in <- e
		} else {
			inactives = append(inactives, c)
		}
	}

	for j := range inactives {
		s.remove(inactives[j])
	}
}

func (s *server) add(conn net.Conn) (*client, error) {
	l := len(s.clients)
	if l >= s.maxClients {
		conn.Close()
		return nil, errors.New("Max connections reached, closing connection.")
	}
	c := newClient(l, conn, s)
	s.clients = append(s.clients, c)
	go c.listen()
	return c, nil
}

func (s *server) remove(c *client) {
	l := len(s.clients) - 1
	i := c.index
	if i < l {
		swap := s.clients[l]
		s.clients[i] = swap
		swap.index = i
	}
	s.clients = s.clients[:l]
}

func initialResponse(req *http.Request) []byte {
	var buf bytes.Buffer
	buf.WriteString(HEADER)
	if origin := req.Header.Get("origin"); origin != "" {
		cors := fmt.Sprintf("Access-Control-Allow-Origin: %s", origin)
		buf.WriteString(cors)
	}
	buf.WriteString(BODY)
	return buf.Bytes()
}
