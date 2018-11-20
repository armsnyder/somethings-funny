package main

import "github.com/gordonklaus/portaudio"

type audioStream interface {
	close() error
	start() error
	stop() error
}

type defaultAudioStream struct { *portaudio.Stream }

func (s *defaultAudioStream) close() error {
	return s.Stream.Close()
}

func (s *defaultAudioStream) start() error {
	return s.Stream.Start()
}

func (s *defaultAudioStream) stop() error {
	return s.Stream.Stop()
}
