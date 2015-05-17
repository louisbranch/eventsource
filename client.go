package eventsource

import (
	"net"
	"time"
)

type client struct {
	active     bool
	index      int
	connection net.Conn
	server     *server
	in         chan event
}

func newClient(index int, conn net.Conn, s *server) *client {
	c := client{}
	c.active = true
	c.index = index
	c.connection = conn
	c.server = s
	c.in = make(chan event, 10)
	return &c
}

func (c *client) write(msg []byte) error {
	_, err := c.connection.Write(msg)
	return err
}

func (c *client) deactivate() {
	c.active = false
	c.connection.Close()
	c.connection = nil
	close(c.in)
}

func (c *client) listen() {
	for {
		e := <-c.in
		c.connection.SetWriteDeadline(time.Now().Add(2 * time.Second))
		err := c.write(e.message)
		if err != nil {
			c.deactivate()
			break
		}
	}
}
