package eventsource

import (
	"bytes"
	"fmt"
	"net/http"
	"strings"
)

const (
	HEADER = `HTTP/1.1 200 OK
Content-Type: text/event-stream
Cache-Control: no-cache
Connection: keep-alive`

	BODY = "retry: 2000\n"

	LIMIT_REACHED = "Max connections reached, closing connection."

	PING = ": ping\n"
)

type Eventsource struct {
	server
}

var padding string

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

func NewServer(maxClients int) *Eventsource {
	e := &Eventsource{
		server{
			limit:  maxClients,
			send:   make(chan event),
			add:    make(chan client),
			remove: make(chan client),
		},
	}
	go e.listen()
	return e
}

func (e *Eventsource) Broadcast(name string, message string, channels []string) {
	event := event{
		name:     name,
		message:  message,
		channels: channels,
	}
	go func() {
		e.send <- event
	}()
}

// ServeHTTP implements the http handle interface.
// If the connection supports hijacking, it sends an initial header to switch
// to text/stream protocol and an initial body to retry after 2 seconds if the
// connection drops.
func (e *Eventsource) ServeHTTP(res http.ResponseWriter, req *http.Request) {
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

	_, err = conn.Write(handshake(req))
	if err != nil {
		conn.Close()
	}

	channels := req.URL.Query().Get("channels")
	c := client{
		conn:     conn,
		channels: strings.Split(channels, ","),
		events:   make(chan []byte),
		done:     make(chan bool),
	}

	e.server.add <- c
}

// handshake sends a header and body sent to client to establish a
// text/stream connection with retry option and CORS enabled.
func handshake(req *http.Request) []byte {
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
