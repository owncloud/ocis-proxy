package proxy

import (
	"github.com/owncloud/ocis-pkg/v2/log"
	"github.com/owncloud/ocis-proxy/pkg/proxy/policy"
	"net/http"
	"net/http/httputil"
)

// RequestRewriterFunc is a function which modifies the request
type RequestRewriterFunc func(r *http.Request) func(r *http.Request)

// MultiHostReverseProxy extends httputil to support multiple hosts with different policies.
type MultiHostReverseProxy struct {
	httputil.ReverseProxy
	logger  log.Logger
	Rewrite RequestRewriterFunc
}

// NewMultiHostReverseProxy undocumented
func NewMultiHostReverseProxy(opts ...Option) *MultiHostReverseProxy {
	options := newOptions(opts...)

	l := options.Logger
	if options.Config.Policies == nil {
		l.Info().Str("source", "runtime").Msg("Policies")
		options.Config.Policies = []policy.Policy{}
	} else {
		l.Info().Str("source", "file").Msg("Policies")
	}

	ps, err := policy.NewStrategy(options.Config.PolicyStrategy)
	if err != nil {
		l.Fatal().Err(err).Msgf("Could not initialize policy-engine")
	}

	pr, err := policy.NewPolicyRewriter(options.Config.Policies, ps, l)
	if err != nil {
		l.Fatal().Err(err).Msgf("Could not initialize policy-engine")
	}

	return &MultiHostReverseProxy{
		Rewrite: pr,
		logger:  l,
	}
}

func (p *MultiHostReverseProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	p.Director = p.Rewrite(r)
	// Call upstream ServeHTTP
	p.ReverseProxy.ServeHTTP(w, r)
}
