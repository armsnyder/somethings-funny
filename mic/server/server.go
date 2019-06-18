package server

import (
	"context"
	"github.com/sirupsen/logrus"
	"net"
	"os"
	"somethings-funny/mic/server/connections"
	"strings"
)

func Serve(ctx context.Context, stream <-chan []byte) error {
	logrus.Info("Opening server socket")

	lis, err := net.Listen("unix", "/mic/mic.sock")
	if err != nil && strings.Contains(err.Error(), "address already in use") {
		if err := os.Remove("/mic/mic.sock"); err != nil {
			return err
		}
		lis, err = net.Listen("unix", "/mic/mic.sock")
	}

	if err != nil {
		return err
	}

	defer func() {
		logrus.Info("Closing server socket")
		_ = lis.Close()
	}()

	conns := connections.New()

	errChan := make(chan error)

	go func() {
		logrus.Info("Server listening for client connections")

		for {
			conn, err := lis.Accept()

			if err != nil {
				select {
				case errChan <- err:
				case <-ctx.Done():
				}

				return
			}

			id := conns.Add(conn)

			logrus.Infof("Client %d connected to server", id)
		}
	}()

	go func() {
		for {
			select {
			case chunk := <-stream:
				var deadConns []uint64

				conns.ForEach(func(conn net.Conn, id uint64) {
					if _, err := conn.Write(chunk); err != nil {
						logrus.Infof("Client %d disconnected from server", id)

						_ = conn.Close()

						deadConns = append(deadConns, id)
					}
				})

				if len(deadConns) > 0 {
					conns.Remove(deadConns...)
				}

			case <-ctx.Done():
				return
			}
		}
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case err = <-errChan:
		return err
	}
}
