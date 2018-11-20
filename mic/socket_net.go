package main

import "net"

type socketNet interface {
	listen(string, string) (socketListener, error)
}

type defaultSocketNet struct {}

func (*defaultSocketNet) listen(network, address string) (socketListener, error) {
	lis, err := net.Listen(network, address)
	return &defaultSocketListener{lis}, err
}
