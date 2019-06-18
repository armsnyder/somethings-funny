package main

import (
	"context"
	"github.com/sirupsen/logrus"
	"os"
	"os/signal"
	"somethings-funny/mic/hardware"
	"somethings-funny/mic/read"
	"somethings-funny/mic/server"
	"sync"
	"syscall"
)

func main() {
	ctx, cancelFunc := context.WithCancel(context.Background())

	errChan := make(chan error)

	wg := new(sync.WaitGroup)

	run := func(runnable func() error) {
		wg.Add(1)
		defer wg.Done()
		select {
		case <-ctx.Done():
		case errChan <- runnable():
		}
	}

	chunks := make(chan []int16, 8)

	rawData := make(chan []byte)

	go run(func() error {
		return hardware.Stream(ctx, chunks)
	})

	go run(func() error {
		return read.Read(ctx, chunks, rawData)
	})

	go run(func() error {
		return server.Serve(ctx, rawData)
	})

	sigIntChan := make(chan os.Signal)

	signal.Notify(sigIntChan, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-sigIntChan:
	case err := <-errChan:
		logrus.Error(err)
	}

	logrus.Info("Beginning graceful shutdown")

	cancelFunc()

	wg.Wait()
}
