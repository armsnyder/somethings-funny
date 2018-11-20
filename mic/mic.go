package main

import (
	"encoding/binary"
	"fmt"
	"github.com/gordonklaus/portaudio"
)

var numFrames = 256

type connectionsContainer struct{ connections []socketConnection }

type mic struct {
	audioInput     audioInput
	socketNet      socketNet
	sigTermChannel <-chan interface{}
}

func (m *mic) start() {
	fmt.Println("Running")
	defer func() {
		fmt.Println("Finished")
	}()
	interrupt := make(chan bool)
	done := make(chan bool)
	connections := connectionsContainer{make([]socketConnection, 0)}
	go m.listenForConnections(interrupt, done, func(conn *socketConnection) {
		fmt.Println("New connection established by client")
		connections.connections = append(connections.connections, *conn)
	})
	defer func() {
		interrupt <- true
		<-done
	}()

	writeBuffer := make([]byte, 2*numFrames)
	go m.listenOnMicrophone(interrupt, done, func(data []int16,
		timeInfo portaudio.StreamCallbackTimeInfo, flags portaudio.StreamCallbackFlags) {

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

	<-m.sigTermChannel
	fmt.Println("Exit signal received")
}

func chk(err error) {
	if err != nil {
		panic(err)
	}
}

func (m *mic) listenForConnections(interrupt <-chan bool, done chan<- bool,
	callback func(conn *socketConnection)) {

	defer func() {
		done <- true
	}()
	fmt.Println("Opening unix socket")
	ln, err := m.socketNet.listen("unix", "/mic/mic.sock")
	closing := false
	defer func() {
		fmt.Println("Closing socket")
		closing = true
		ln.close()
	}()
	chk(err)
	go func() {
		fmt.Println("Listening for client connections")
		for {
			conn, err := ln.accept()
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

func (m *mic) listenOnMicrophone(interrupt <-chan bool, done chan<- bool,
	callback func([]int16, portaudio.StreamCallbackTimeInfo, portaudio.StreamCallbackFlags)) {

	defer func(done chan<- bool) {
		done <- true
	}(done)
	fmt.Println("Initializing portaudio")
	chk(m.audioInput.initialize())
	defer func() {
		fmt.Println("Terminating portaudio")
		chk(m.audioInput.terminate())
	}()
	info, err := m.audioInput.defaultInputDevice()
	chk(err)
	fmt.Println("Device info", info)
	fmt.Println("Opening audio stream")
	stream, err := m.audioInput.openDefaultStream(1, 0, 44100, numFrames, callback)
	chk(err)
	defer func() {
		fmt.Println("Closing audio stream")
		chk(stream.close())
	}()
	fmt.Println("Starting audio stream")
	chk(stream.start())
	defer func() {
		fmt.Println("Stopping audio stream")
		chk(stream.stop())
	}()
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
		_, err := conn.write(*writeBuffer)
		if err != nil {
			fmt.Println("Write failed. Closing client connection.")
			conn.close()
			connections.connections =
				append(connections.connections[:i], connections.connections[i+1:]...)
		}
	}
}
