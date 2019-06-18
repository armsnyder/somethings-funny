package instrument

import (
	"context"
	"github.com/envoyproxy/go-control-plane/envoy/api/v2"
	"github.com/envoyproxy/go-control-plane/pkg/server"
	"github.com/sirupsen/logrus"
)

type xdsLogger struct {
}

func (*xdsLogger) OnStreamOpen(ctx context.Context, i int64, typ string) error {
	logrus.Infof("OnStreamOpen %d %s", i, typ)
	return nil
}

func (*xdsLogger) OnStreamClosed(i int64) {
	logrus.Infof("OnStreamClosed %d", i)
}

func (*xdsLogger) OnStreamRequest(i int64, req *v2.DiscoveryRequest) error {
	logrus.Infof("OnStreamRequest %d %s %s", i, req.TypeUrl, req.VersionInfo)
	return nil
}

func (*xdsLogger) OnStreamResponse(i int64, req *v2.DiscoveryRequest, res *v2.DiscoveryResponse) {
	logrus.Infof("OnStreamResponse %d %s %s", i, res.TypeUrl, res.VersionInfo)
}

func (*xdsLogger) OnFetchRequest(context.Context, *v2.DiscoveryRequest) error {
	return nil
}

func (*xdsLogger) OnFetchResponse(*v2.DiscoveryRequest, *v2.DiscoveryResponse) {
}

func NewXdsLogger() server.Callbacks {
	return &xdsLogger{}
}
