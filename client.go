package eventsource

import (
	"net"
)

type client struct {
	active     bool
	index      uint
	connection net.Conn
	server     *server
}

func newClient(index uint, conn net.Conn, s *server) *client {
	c := client{}
	c.active = true
	c.index = index
	c.connection = conn
	c.server = s
	return &c
}

func (c *client) write(msg []byte) error {
	_, err := c.connection.Write(msg)
	return err
}
