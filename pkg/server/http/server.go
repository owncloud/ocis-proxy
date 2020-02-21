package http

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/http/httputil"
	"net/url"

	"github.com/owncloud/ocis-bridge/pkg/version"
	ocCrypto "github.com/owncloud/ocis-konnectd/pkg/crypto"
	svc "github.com/owncloud/ocis-pkg/v2/service/http"
)

type proxyHandler struct {
	handler *httputil.ReverseProxy
}

// Server initializes the http service and server.
func Server(opts ...Option) (svc.Service, error) {
	options := newOptions(opts...)

	if err := ocCrypto.GenCert(options.Logger); err != nil {
		log.Fatal(err)
	}

	cer, err := tls.LoadX509KeyPair("server.crt", "server.key")
	if err != nil {
		log.Fatal(err)
	}

	tlsConf := tls.Config{
		Certificates: []tls.Certificate{cer},
	}

	service := svc.NewService(
		svc.Logger(options.Logger),
		svc.Namespace(options.Namespace),
		svc.Name("web.bridge"),
		svc.Version(version.String),
		svc.Address(options.Config.HTTP.Addr),
		svc.Context(options.Context),
		svc.Flags(options.Flags...),
		svc.TLSConfig(&tlsConf),
	)

	service.Handle("/", handler(phoenix))
	service.Handle("/ocs/", handler(reva))
	service.Handle("/signin/", handler(konnectd))
	service.Handle("/konnect/", handler(konnectd))
	service.Handle("/remote.php/webdav/", handler(reva))
	service.Handle("/.well-known/openid-configuration", handler(konnectd))

	service.Init()
	return service, nil
}

func mustURL(s string) *url.URL {
	url, err := url.Parse(s)
	if err != nil {
		log.Fatal(err)
	}

	return url
}

func handler(port servicePort) *httputil.ReverseProxy {
	return httputil.NewSingleHostReverseProxy(mustURL(fmt.Sprintf("http://localhost:%v", port)))
}

type servicePort string

var (
	phoenix  servicePort = "9100"
	konnectd servicePort = "9130"
	reva     servicePort = "9140"
	webdav   servicePort = "9115"
)
