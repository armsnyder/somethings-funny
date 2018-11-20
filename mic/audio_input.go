package main

import "github.com/gordonklaus/portaudio"

type audioInput interface {
	initialize() error
	terminate() error
	defaultInputDevice() (interface{}, error)
	openDefaultStream(int, int, float64, int, func([]int16, portaudio.StreamCallbackTimeInfo,
		portaudio.StreamCallbackFlags)) (audioStream, error)
}

type defaultAudioInput struct {}

func (*defaultAudioInput) initialize() error {
	return portaudio.Initialize()
}

func (*defaultAudioInput) terminate() error {
	return portaudio.Terminate()
}

func (*defaultAudioInput) defaultInputDevice() (interface{}, error) {
	return portaudio.DefaultInputDevice()
}

func (*defaultAudioInput) openDefaultStream(numInputChannels, numOutputChannels int,
	sampleRate float64, framesPerBuffer int,
	callback func([]int16, portaudio.StreamCallbackTimeInfo,
		portaudio.StreamCallbackFlags)) (audioStream, error) {

	s, err := portaudio.OpenDefaultStream(numInputChannels, numOutputChannels, sampleRate,
		framesPerBuffer, callback)
	return &defaultAudioStream{s}, err
}
