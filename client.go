package eventsource

import (
	"net"
	"time"
)

type client struct {
	active   bool
	index    int
	conn     net.Conn
	in       chan Event
	channels []string
}

func newClient(index int, conn net.Conn, channels []string) *client {
	c := client{}
	c.active = true
	c.index = index
	c.conn = conn
	c.channels = channels
	c.in = make(chan Event, 10)
	return &c
}

func (c *client) write(msg []byte) error {
	_, err := c.conn.Write(msg)
	return err
}

func (c *client) deactivate() {
	c.active = false
	c.conn.Close()
	c.conn = nil
}

func (c *client) listen() {
loop:
	for {
		e := <-c.in
		c.conn.SetWriteDeadline(time.Now().Add(2 * time.Second))
		err := c.write(e.Bytes())
		if err != nil {
			c.deactivate()
			break loop
		}
	}
}
