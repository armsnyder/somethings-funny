package main

import "net"

type socketListener interface {
	close() error
	accept() (socketConnection, error)
}

type defaultSocketListener struct { net.Listener }

func (lis *defaultSocketListener) close() error {
	return lis.Listener.Close()
}

func (lis *defaultSocketListener) accept() (socketConnection, error) {
	conn, err := lis.Listener.Accept()
	return &defaultSocketConnection{conn}, err
}
