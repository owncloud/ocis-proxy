package http

import (
	"crypto/tls"
	"net/http"
	"os"

	occrypto "github.com/owncloud/ocis-konnectd/pkg/crypto"
	svc "github.com/owncloud/ocis-pkg/v2/service/http"
	"github.com/owncloud/ocis-proxy/pkg/version"
)

// Server initializes the http service and server.
func Server(opts ...Option) (svc.Service, error) {
	options := newOptions(opts...)

	// GenCert has side effects as it writes 2 files to the binary running location
	occrypto.GenCert(options.Logger)

	cer, err := tls.LoadX509KeyPair("localhost.pem", "localhost-key.pem")
	if err != nil {
		options.Logger.Fatal().Err(err).Msg("Could not setup TLS")
		os.Exit(1)
	}

	config := &tls.Config{Certificates: []tls.Certificate{cer}}
	service := svc.NewService(
		svc.Name("web.proxy"),
		svc.TLSConfig(config),
		svc.Logger(options.Logger),
		svc.Namespace(options.Namespace),
		svc.Version(version.String),
		svc.Address(options.Config.HTTP.Addr),
		svc.Context(options.Context),
		svc.Flags(options.Flags...),
		svc.Handler(applyMiddlewares(
			options.Handler,
			options.Middlewares...),
		),
	)

	service.Init()

	return service, nil
}

func applyMiddlewares(h http.Handler, mws ...func(handler http.Handler) http.Handler) http.Handler {
	var han = h
	for _, mw := range mws {
		han = mw(han)
	}

	return han
}
