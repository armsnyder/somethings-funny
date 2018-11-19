package main

import (
	"encoding/binary"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/gordonklaus/portaudio"
)

var numFrames = 256

type connectionsContainer struct{ connections []net.Conn }

func chk(err error) {
	if err != nil {
		panic(err)
	}
}

func listenForConnections(interrupt <-chan bool, done chan<- bool, callback func(conn *net.Conn)) {
	defer func() {
		done <- true
	}()
	fmt.Println("Opening unix socket")
	ln, err := net.Listen("unix", "/mic/mic.sock")
	closing := false
	defer func() {
		fmt.Println("Closing socket")
		closing = true
		ln.Close()
	}()
	chk(err)
	go func() {
		fmt.Println("Listening for client connections")
		for {
			conn, err := ln.Accept()
			if err != nil {
				if closing {
					return
				}
				panic(err)
			}
			callback(&conn)
		}
	}()
	<-interrupt
}

func listenOnMicrophone(interrupt <-chan bool, done chan<- bool, callback func([]int16, portaudio.StreamCallbackTimeInfo, portaudio.StreamCallbackFlags)) {
	defer func() {
		done <- true
	}()
	fmt.Println("Initializing portaudio")
	portaudio.Initialize()
	defer func() {
		fmt.Println("Terminating portaudio")
		chk(portaudio.Terminate())
	}()
	info, err := portaudio.DefaultInputDevice()
	chk(err)
	fmt.Println("Device info", *info)
	fmt.Println("Opening audio stream")
	stream, err := portaudio.OpenDefaultStream(1, 0, 44100, numFrames, callback)
	chk(err)
	defer func(stream *portaudio.Stream) {
		fmt.Println("Closing audio stream")
		chk(stream.Close())
	}(stream)
	fmt.Println("Starting audio stream")
	chk(stream.Start())
	defer func(stream *portaudio.Stream) {
		fmt.Println("Stopping audio stream")
		chk(stream.Stop())
	}(stream)
	<-interrupt
}

func intsToBytes(intBuffer *[]int16, byteBuffer *[]byte) {
	if len(*byteBuffer) < len(*intBuffer)*2 {
		panic("Not enough room in byteBuffer")
	}
	for i, v := range *intBuffer {
		binary.BigEndian.PutUint16((*byteBuffer)[i*2:(i+1)*2], uint16(v))
	}
}

func writeToConnections(connections *connectionsContainer, writeBuffer *[]byte) {
	for i := 0; i < len(connections.connections); i++ {
		conn := connections.connections[i]
		_, err := conn.Write(*writeBuffer)
		if err != nil {
			fmt.Println("Write failed. Closing client connection.")
			conn.Close()
			connections.connections = append(connections.connections[:i], connections.connections[i+1:]...)
		}
	}
}

func main() {
	fmt.Println("Running")
	defer func() {
		fmt.Println("Finished")
	}()
	interrupt := make(chan bool)
	done := make(chan bool)
	connections := connectionsContainer{make([]net.Conn, 0)}
	go listenForConnections(interrupt, done, func(conn *net.Conn) {
		fmt.Println("New connection established by client")
		connections.connections = append(connections.connections, *conn)
	})
	defer func() {
		interrupt <- true
		<-done
	}()

	writeBuffer := make([]byte, 2*numFrames)
	go listenOnMicrophone(interrupt, done, func(data []int16, timeInfo portaudio.StreamCallbackTimeInfo, flags portaudio.StreamCallbackFlags) {
		if flags != 0 {
			fmt.Println("flags in callback", flags)
		}
		if len(connections.connections) == 0 {
			return
		}
		intsToBytes(&data, &writeBuffer)
		writeToConnections(&connections, &writeBuffer)
	})
	defer func() {
		interrupt <- true
		<-done
	}()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGTERM)
	<-sig
	fmt.Println("Exit signal received")
}
