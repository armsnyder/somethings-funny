package instrument

import (
	"github.com/sirupsen/logrus"
	"net/http"
	"net/http/httputil"
)

type incomingRequestLogger struct {
	delegate http.Handler
}

func (this *incomingRequestLogger) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	if logrus.IsLevelEnabled(logrus.DebugLevel) {
		writer := &capturingWriter{
			delegate:   res,
			statusCode: 200,
		}
		this.delegate.ServeHTTP(writer, req)
		dumpedReq, _ := httputil.DumpRequest(req, false)
		logrus.WithField("request", string(dumpedReq)).Infof(
			"Incoming [%d] %s", writer.statusCode, req.URL)
	} else {
		this.delegate.ServeHTTP(res, req)
	}
}

func NewLoggingHttpHandler(delegate http.Handler) http.Handler {
	return &incomingRequestLogger{
		delegate: delegate,
	}
}

type capturingWriter struct {
	delegate   http.ResponseWriter
	statusCode int
}

func (this *capturingWriter) Header() http.Header {
	return this.delegate.Header()
}

func (this *capturingWriter) Write(b []byte) (int, error) {
	return this.delegate.Write(b)
}

func (this *capturingWriter) WriteHeader(statusCode int) {
	this.statusCode = statusCode
	this.delegate.WriteHeader(statusCode)
}
