package main

import (
	"os"
	"os/signal"
	"syscall"
)

func sigTermChannel() <-chan interface{} {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGTERM)
	result := make(chan interface{}, 1)
	go func(sig chan os.Signal, result chan interface{}) {
		<-sig
		result <- true
	}(sig, result)
	return result
}
