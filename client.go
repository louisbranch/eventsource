package eventsource

import (
	"net"
	"time"
)

type client struct {
	active   bool
	index    int
	channels []string
	conn     net.Conn
	in       chan job
}

func newClient(index int, conn net.Conn, channels []string) *client {
	c := client{}
	c.active = true
	c.index = index
	c.conn = conn
	c.channels = channels
	c.in = make(chan job)
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

func (c *client) loop() {
	for {
		p, ok := <-c.in
		if ok {
			c.conn.SetWriteDeadline(time.Now().Add(2 * time.Second))
			err := c.write(p.data)
			if err == nil {
				p.done <- true
			} else {
				c.deactivate()
				p.done <- false
			}
		} else {
			return
		}
	}
}
