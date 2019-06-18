package instrument

import (
	"github.com/sirupsen/logrus"
	"net/http"
	"net/http/httputil"
)

type loggingTransport struct {
	delegate http.RoundTripper
}

func (l *loggingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	response, err := l.delegate.RoundTrip(req)
	reqBytes, _ := httputil.DumpRequestOut(req, false)
	respBytes, _ := httputil.DumpResponse(response, true)
	logrus.WithField(
		"request", string(reqBytes)).WithField(
		"response", string(respBytes)).Infof(
		"Outgoing [%s] %s", response.Status, req.URL)
	return response, err
}

func NewLoggingHttpClient() *http.Client {
	return &http.Client{
		Transport: &loggingTransport{
			http.DefaultTransport,
		},
	}
}
