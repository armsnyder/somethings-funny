package main

import (
	"context"
	"github.com/kelseyhightower/envconfig"
	"github.com/sirupsen/logrus"
	"log"
	"os"
	"os/signal"
	"somethings-funny/sampler/filter"
	"somethings-funny/sampler/read"
	"somethings-funny/sampler/upload"
	"sync"
	"syscall"
)

type envSpec struct {
	UploadUrl string  `envconfig:"upload_url" required:"true"`
	Token     string  `envconfig:"token" required:"true"`
	Threshold float64 `envconfig:"threshold" required:"true"`
}

var env envSpec

func init() {
	err := envconfig.Process("", &env)
	if err != nil {
		log.Fatal(err.Error())
	}
}

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

	socketReader := make(chan []byte, 8)

	aggregateData := make(chan *filter.WavFile, 8)

	go run(func() error {
		return read.Read(ctx, socketReader)
	})

	go run(func() error {
		return filter.AggregateFilter(ctx, env.Threshold, socketReader, aggregateData)
	})

	go run(func() error {
		return upload.Upload(ctx, env.UploadUrl, env.Token, aggregateData)
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
