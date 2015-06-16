package eventsource

import "net/http"

// An Eventsource is a high-level server abstraction. It can be used as a
// Handler for a http route and to send events to clients. Multiple servers can
// coexist and be used on more than one end-point.
type Eventsource struct {
	server

	// Interface that implements how channels are assigned to clients. It
	// defaults to NoChannels, meaning all events must be global.
	ChanSub ChannelSubscriber

	// Interface that implements what options are sent during the initial http
	// handshaking. See DefaultHttpOptions for built-in options.
	HttpOptions HttpOptions

	// Interface that implements how metrics are tracked. Defaults to
	// StatsJSONLogger
	Stats Stats
}

// A HijackingError is displayed when the browser doesn't support connection
// hijacking. See http://golang.org/pkg/net/http/#Hijacker
var HijackingError = "webserver doesn't support hijacking"

// Global stats tracking
var stats Stats

// The Start function sets all undefined options to their defaults and configure
// the underlining server to start listening to events
func (es *Eventsource) Start() {
	if es.ChanSub == nil {
		es.ChanSub = NoChannels{}
	}

	if es.HttpOptions == nil {
		es.HttpOptions = DefaultHttpOptions{
			Retry:             2000,
			Cors:              true,
			OldBrowserSupport: true,
		}
	}

	if es.Stats == nil {
		es.Stats = &StatsJSONLogger{}
	}
	// push to global state
	stats = es.Stats

	es.server = server{
		add:    make(chan client),
		remove: make(chan client),
		local:  make(chan Event),
		global: make(chan Event),
	}

	go es.server.listen()
}

// The send function sends an event to all clients that have  subscribed to one
// of the channels passed.
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
// If the connection supports hijacking, it sends an initial header and body to
// switch to the text/stream protocol and start streaming.
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
		events:   make(chan payload),
		done:     make(chan bool),
	}

	e.server.add <- c
}
