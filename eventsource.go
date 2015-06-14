/*
Package eventsource provides an implementation to Server-sent events using
goroutines to handle client (un)subscription and forward events to clients.
For more information about Eventsource / SSE check the MDN documentation:
https://developer.mozilla.org/en-US/docs/Server-sent_events/Using_server-sent_events
*/
package eventsource

import (
	"bytes"
	"fmt"
	"net/http"
)

const (
	// HTTP HEADER sent to browser to upgrade the protocol to event-stream
	HEADER = `HTTP/1.1 200 OK
Content-Type: text/event-stream
Cache-Control: no-cache
Connection: keep-alive`

	// BODY is the initial payload sent by the server and informs the client to
	// retry a new connection after 2 seconds if it drops.
	BODY = "retry: 2000\n"

	// PING is a message sent every 30s to detect stale clients and remove them
	// from the list.
	PING = ": ping\n"
)

// An Eventsource is a high-level server abstraction. It can be used as a
// Handler for a http route and to send events to clients. An Eventsource
// instance MUST be created using the NewServer function. Multiple servers can
// coexist and be used on more than one end-point.
type Eventsource struct {
	server
	ChanSub ChannelSubscriber
}

// Internet Explorer < 10 needs a message padding to successfully establish a
// text stream connection See
//http://blogs.msdn.com/b/ieinternals/archive/2010/04/06/comet-streaming-in-internet-explorer-with-xmlhttprequest-and-xdomainrequest.aspx
var padding string

func init() {
	var buf bytes.Buffer
	buf.WriteByte(':')
	for i := 0; i < 2048; i++ {
		buf.WriteByte(' ')
	}
	buf.WriteByte('\n')
	padding = buf.String()
}

// The NewServer function configures a new instace of the Eventsource, creating
// all necessary channels and spawning a new goroutine to listen to commands.
func NewServer() *Eventsource {
	e := &Eventsource{
		server: server{
			add:    make(chan client),
			remove: make(chan client),
			local:  make(chan Event),
			global: make(chan Event),
		},
		ChanSub: NoChannels{},
	}
	go e.listen()
	return e
}

// The send function sends an event to all clients that have
// subscribed to one of the channels passed.
func (e *Eventsource) Send(event Event) {
	go func() {
		e.local <- event
	}()
}

// The broadcast function sends an event to all clients.
func (e *Eventsource) Broadcast(event Event) {
	go func() {
		e.global <- event
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

	channels := e.ChanSub.ParseRequest(req)

	c := client{
		conn:     conn,
		channels: channels,
		events:   make(chan []byte),
		done:     make(chan bool),
	}

	e.server.add <- c
}

// The handshake function sends a header and body to the browser to establish a
// text/stream connection with retry option and CORS enabled.
func handshake(req *http.Request) []byte {
	var buf bytes.Buffer
	buf.WriteString(HEADER)
	if origin := req.Header.Get("origin"); origin != "" {
		cors := fmt.Sprintf("Access-Control-Allow-Origin: %s\n", origin)
		buf.WriteString("Access-Control-Allow-Credentials: true\n")
		buf.WriteString(cors)
	}
	buf.WriteString("\n\n")
	buf.WriteString(padding)
	buf.WriteString(BODY)
	return buf.Bytes()
}
