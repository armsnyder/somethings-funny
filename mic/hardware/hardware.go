package hardware

import (
	"context"
	"github.com/gordonklaus/portaudio"
	"github.com/sirupsen/logrus"
	"somethings-funny/common"
)

func Stream(ctx context.Context, output chan<- []int16) error {

	logrus.Info("Initializing portaudio")

	err := portaudio.Initialize()

	if err != nil {
		return err
	}

	device, err := portaudio.DefaultInputDevice()

	if err != nil {
		return err
	}

	logrus.Infof("Default device: %+v", device)

	defer func() {
		logrus.Info("Terminating portaudio")
		_ = portaudio.Terminate()
	}()

	logrus.Info("Opening portaudio stream")

	buff := make([]int16, common.FramesPerBuffer)
	stream, err := portaudio.OpenDefaultStream(1, 0, float64(common.SampleRate),
		common.FramesPerBuffer, &buff)

	if err != nil {
		return err
	}

	defer func() {
		logrus.Info("Closing portaudio stream")
		_ = stream.Close()
	}()

	logrus.Info("Starting portaudio stream")

	if err = stream.Start(); err != nil {
		return err
	}

	errChan := make(chan error)
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				err := stream.Read()
				if err != nil {
					if err == portaudio.InputOverflowed {
						logrus.Warn(err.Error())
					} else {
						errChan <- err
						return
					}
				}
				cp := make([]int16, len(buff))
				for i, b := range buff {
					cp[i] = b
				}
				output <- cp
			}
		}
	}()

	defer func() {
		logrus.Info("Stopping portaudio stream")
		_ = stream.Stop()
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-errChan:
		return err
	}
}
