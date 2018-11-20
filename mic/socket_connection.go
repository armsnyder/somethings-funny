package main

import "net"

type socketConnection interface {
	write([]byte) (int, error)
	close() error
}

type defaultSocketConnection struct { net.Conn }

func (c *defaultSocketConnection) write(b []byte) (int, error) {
	return c.Conn.Write(b)
}

func (c *defaultSocketConnection) close() error {
	return c.Conn.Close()
}
