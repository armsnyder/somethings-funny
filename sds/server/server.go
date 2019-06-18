package server

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"github.com/cenkalti/backoff"
	"github.com/envoyproxy/go-control-plane/envoy/api/v2/auth"
	"github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
	discovery "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v2"
	"github.com/envoyproxy/go-control-plane/pkg/cache"
	xds "github.com/envoyproxy/go-control-plane/pkg/server"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/acme"
	"golang.org/x/crypto/acme/autocert"
	"google.golang.org/grpc"
	"math/rand"
	"net"
	"net/http"
	"os"
	"os/signal"
	"somethings-funny/sds/instrument"
	"time"
)

type hash struct{}

const (
	monoId   = "mono"
	acmePort = 30328
	grpcPort = 10962
)

func (hash) ID(node *core.Node) string {
	return monoId
}

func Run(certCache autocert.Cache, letsEncryptUrl, domain string) {
	certManager := &autocert.Manager{
		Cache:      certCache,
		Prompt:     autocert.AcceptTOS,
		HostPolicy: autocert.HostWhitelist(domain),
		Client: &acme.Client{
			HTTPClient:   instrument.NewLoggingHttpClient(),
			DirectoryURL: letsEncryptUrl,
		},
	}

	go serveAcme(certManager)

	snapshotCache := cache.NewSnapshotCache(false, hash{}, nil)

	err := snapshotCache.SetSnapshot(monoId, makeSnapshot([]byte{}, []byte{}))
	if err != nil {
		logrus.WithError(err).Fatal("Failed to set initial snapshot")
	}

	go serveGrpc(snapshotCache)

	go doBackgroundCertRefresh(certManager, snapshotCache, domain)

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt)
	<-c
}

func serveAcme(certManager *autocert.Manager) {
	logrus.Infof("Listening for Let's Encrypt ACME challenges on port %d", acmePort)
	err := http.ListenAndServe(fmt.Sprintf(":%d", acmePort),
		instrument.NewLoggingHttpHandler(certManager.HTTPHandler(nil)))
	if err != nil {
		logrus.WithError(err).Fatal("ACME challenge listener exited")
	}
}

func serveGrpc(snapshotCache cache.SnapshotCache) {
	lis, _ := net.Listen("tcp", fmt.Sprintf(":%d", grpcPort))
	server := xds.NewServer(snapshotCache, instrument.NewXdsLogger())
	grpcServer := grpc.NewServer()
	discovery.RegisterSecretDiscoveryServiceServer(grpcServer, server)
	logrus.Infof("Listening for Envoy XDS requests on port %d", grpcPort)
	if err := grpcServer.Serve(lis); err != nil {
		logrus.WithError(err).Fatal("XDS listener exited")
	}
}

func doBackgroundCertRefresh(certManager *autocert.Manager, snapshotCache cache.SnapshotCache,
	domain string) {

	var lastCert [][]byte
	var crt *tls.Certificate
	for {
		attempt := 0
		err := backoff.Retry(func() (err error) {
			attempt++
			crt, err = certManager.GetCertificate(&tls.ClientHelloInfo{ServerName: domain})
			if err != nil {
				logrus.WithError(err).WithField("attempt", attempt).Debug(
					"Failed to get certificate from autocert manager. Retrying.")
			}
			return
		}, backoff.NewExponentialBackOff())
		if err != nil {
			logrus.WithError(err).Fatal(
				"Failed to get certificate from autocert manager after many retries")
		}

		if !areCertChainsEqual(lastCert, crt.Certificate) {
			certPem, err := pemBlockForCertChain(crt.Certificate)
			if err != nil {
				logrus.WithError(err).Fatal("Failed to get certificate as PEM")
			}

			pvtKeyPem, err := pemBlockForKey(crt.PrivateKey)
			if err != nil {
				logrus.WithError(err).Fatal("Failed to get private key as PEM")
			}

			snapshot := makeSnapshot(certPem, pvtKeyPem)
			if err := snapshotCache.SetSnapshot(monoId, snapshot); err != nil {
				logrus.WithError(err).Fatal("Failed to update snapshot")
			}
		}

		day := time.Hour * 24
		jitter := time.Duration(rand.Int31n(60)) * time.Minute
		time.Sleep(day + jitter)
	}
}

func pemBlockForKey(priv interface{}) ([]byte, error) {
	switch k := priv.(type) {
	case *rsa.PrivateKey:
		return pem.EncodeToMemory(
			&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(k)}), nil
	case *ecdsa.PrivateKey:
		b, err := x509.MarshalECPrivateKey(k)
		if err != nil {
			return nil, errors.WithMessagef(err, "Unable to marshal ECDSA private key")
		}
		return pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: b}), nil
	default:
		return nil, errors.New(fmt.Sprintf("unrecognized key type: %T", priv))
	}
}

func pemBlockForCertChain(chain [][]byte) ([]byte, error) {
	certChainBuf := new(bytes.Buffer)
	for _, b := range chain {
		err := pem.Encode(certChainBuf, &pem.Block{
			Type:  "CERTIFICATE",
			Bytes: b,
		})
		if err != nil {
			return nil, err
		}
	}
	return certChainBuf.Bytes(), nil
}

func areCertChainsEqual(a, b [][]byte) bool {
	if a == nil || b == nil || len(a) != len(b) {
		return false
	}
	for i := 0; i < len(a); i++ {
		if bytes.Compare(a[i], b[i]) != 0 {
			return false
		}
	}
	return true
}

func makeSnapshot(certChain, privateKey []byte) cache.Snapshot {
	return cache.Snapshot{
		Secrets: cache.NewResources(
			fmt.Sprint(time.Now().Unix()),
			[]cache.Resource{
				&auth.Secret{
					Name: "cert",
					Type: &auth.Secret_TlsCertificate{
						TlsCertificate: &auth.TlsCertificate{
							CertificateChain: &core.DataSource{
								Specifier: &core.DataSource_InlineBytes{
									InlineBytes: certChain,
								},
							},
							PrivateKey: &core.DataSource{
								Specifier: &core.DataSource_InlineBytes{
									InlineBytes: privateKey,
								},
							},
						},
					},
				},
			},
		),
	}
}
