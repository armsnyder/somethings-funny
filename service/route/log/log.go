package log

import (
	"github.com/sirupsen/logrus"
	"io"
)

type fnWriter struct {
	fn func(...interface{})
}

func (this *fnWriter) Write(p []byte) (n int, err error) {
	this.fn(string(p))
	return len(p), nil
}

func NewInfoWriter() io.Writer {
	return &fnWriter{fn: logrus.Info}
}

func NewErrorWriter() io.Writer {
	return &fnWriter{fn: logrus.Error}
}
