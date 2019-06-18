package read

import (
	"context"
	"github.com/sirupsen/logrus"
	"net"
	"somethings-funny/common"
)

const (
	bitsPerByte = 8
)

func Read(ctx context.Context, reader chan<- []byte) error {
	logrus.Info("Dialing unix socket")
	conn, err := net.Dial("unix", "/mic/mic.sock")
	if err != nil {
		return err
	}
	defer func() {
		logrus.Info("Closing unix socket connection")
		_ = conn.Close()
	}()

	logrus.Info("Listening for audio on unix socket")
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		bytesPerSample := common.BitDepth / bitsPerByte
		readBuffer := make([]byte, common.FramesPerBuffer*bytesPerSample)
		n, err := conn.Read(readBuffer)
		if err != nil {
			return err
		}
		reader <- readBuffer[:n]
	}
}
