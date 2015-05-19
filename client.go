package eventsource

import (
	"net"
	"time"
)

type client struct {
	active     bool
	index      int
	connection net.Conn
	server     *Server
	in         chan event
}

func newClient(index int, conn net.Conn, s *Server) *client {
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
}

func (c *client) listen() {
loop:
	for {
		e := <-c.in
		c.connection.SetWriteDeadline(time.Now().Add(2 * time.Second))
		err := c.write(e.Bytes())
		if err != nil {
			c.deactivate()
			break loop
		}
	}
}
