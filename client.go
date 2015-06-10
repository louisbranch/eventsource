package eventsource

import (
	"net"
	"time"
)

// A client hold the actual connection to the browser, the channels names the
// client has subscribed to, a queue to receive events and a done channel for
// syncronization with pending events.
type client struct {
	events   chan []byte
	done     chan bool
	channels []string
	conn     net.Conn
}

// The listen function receives incoming events on the events channel, writing
// them to its underlining connection. If there is an error, the client send a
// message to remove itself from the pool through the remove channel passed in
// and notifies pending events through closing the done channel.
func (c *client) listen(remove chan<- client) {
	for {
		e, ok := <-c.events
		if !ok {
			c.conn.Close()
			return
		}
		c.conn.SetWriteDeadline(time.Now().Add(2 * time.Second))
		_, err := c.conn.Write(e)
		if err != nil {
			remove <- *c
			c.conn.Close()
			close(c.done)
			break
		}
	}
}
