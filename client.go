package eventsource

import (
	"net"
	"time"
)

type client struct {
	events   chan []byte
	done     chan bool
	channels []string
	conn     net.Conn
}

func (c *client) listen(remove chan<- client) {
	for {
		e, ok := <-c.events
		if !ok {
			c.conn.Close()
			return
		}
		c.conn.SetWriteDeadline(time.Now().Add(2 * time.Second))
		err := c.write(e)
		if err != nil {
			remove <- *c
			c.conn.Close()
			close(c.done)
		}
	}
}

func (c *client) write(msg []byte) error {
	_, err := c.conn.Write(msg)
	return err
}
