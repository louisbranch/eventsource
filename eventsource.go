/*
Package eventsource provides an implementation to Server-sent events using
goroutines to handle client (un)subscription and forward events to clients.
For more information about Eventsource / SSE check the MDN documentation:
https://developer.mozilla.org/en-US/docs/Server-sent_events/Using_server-sent_events
*/
package eventsource

import "net/http"

// An Eventsource is a high-level server abstraction. It can be used as a
// Handler for a http route and to send events to clients. An Eventsource
// instance MUST be created using the NewServer function. Multiple servers can
// coexist and be used on more than one end-point.
type Eventsource struct {
	server

	// Interface that implements how channels are assigned to clients. It
	// defaults to NoChannels, meaning all events must be global.
	ChanSub     ChannelSubscriber
	HttpOptions HttpOptions
}

// A HijackingError is displayed when the browser doesn't support connection
// hijacking. See http://golang.org/pkg/net/http/#Hijacker
var HijackingError = "webserver doesn't support hijacking"

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
		HttpOptions: DefaultHttpOptions{
			Retry:             2000,
			Cors:              true,
			OldBrowserSupport: true,
		},
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
		http.Error(res, HijackingError, http.StatusInternalServerError)
		return
	}

	conn, _, err := hj.Hijack()
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}

	options := e.HttpOptions.Bytes(req)
	_, err = conn.Write(options)
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
