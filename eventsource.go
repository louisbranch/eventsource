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

	MAX_CLIENTS = 1
)

type server struct {
	clients [MAX_CLIENTS]*client
	next    uint
}

func New() *server {
	s := server{}
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
		s.remove(client)
	}
}

func (s *server) add(conn net.Conn) (*client, error) {
	if s.next >= MAX_CLIENTS {
		conn.Close()
		return nil, errors.New("Max connections reached, closing connection.")
	}
	c := newClient(s.next, conn, s)
	s.clients[s.next] = c
	s.next++
	return c, nil
}

func (s *server) remove(c *client) {
	last := s.next - 1
	index := c.index
	if index == last {
		s.next = last
	} else {
		swap := s.clients[last]
		s.clients[index] = swap
		swap.index = index
	}
	c.connection.Close()
	c.connection = nil
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
