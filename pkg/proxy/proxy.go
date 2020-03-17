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
	rp := &MultiHostReverseProxy{
		Directors:      make(map[string]map[string]func(req *http.Request)),
		PolicyStrategy: Migration(),
		logger:         options.Logger,
	}

	l := options.Logger
	ps, err := policy.NewStrategy(options.Config.PolicyStrategy)
	if err != nil {
		l.Fatal().Err(err).Msgf("Could not initialize policy-engine")
	}

	pr, err := policy.NewPolicyRewriter(options.Config.Policies, ps, l)
	if err != nil {
		l.Fatal().Err(err).Msgf("Could not initialize policy-engine")
	}
	if options.Config.Policies == nil {
		rp.logger.Info().Str("source", "runtime").Msg("Policies")
		options.Config.Policies = defaultPolicies()
	} else {
		rp.logger.Info().Str("source", "file").Msg("Policies")
	}

	for _, policy := range options.Config.Policies {
		for _, route := range policy.Routes {
			rp.logger.Debug().Str("fwd: ", route.Endpoint)
			uri, err := url.Parse(route.Backend)
			if err != nil {
				rp.logger.
					Fatal().
					Err(err).
					Msgf("malformed url: %v", route.Backend)
			}

	return &MultiHostReverseProxy{
		Rewrite: pr,
		logger:  l,
			rp.logger.
				Debug().
				Interface("route", route).
				Msg("adding route")

			rp.AddHost(policy.Name, uri, route)
		}
	}

	return rp
}

func singleJoiningSlash(a, b string) string {
	aslash := strings.HasSuffix(a, "/")
	bslash := strings.HasPrefix(b, "/")
	switch {
	case aslash && bslash:
		return a + b[1:]
	case !aslash && !bslash:
		return a + "/" + b
	}
	return a + b
}

// AddHost undocumented
func (p *MultiHostReverseProxy) AddHost(policy string, target *url.URL, rt config.Route) {
	targetQuery := target.RawQuery
	if p.Directors[policy] == nil {
		p.Directors[policy] = make(map[string]func(req *http.Request))
	}
	p.Directors[policy][rt.Endpoint] = func(req *http.Request) {
		req.URL.Scheme = target.Scheme
		req.URL.Host = target.Host
		// Apache deployments host addresses need to match on req.Host and req.URL.Host
		// see https://stackoverflow.com/questions/34745654/golang-reverseproxy-with-apache2-sni-hostname-error
		if rt.ApacheVHost {
			req.Host = target.Host
		}

		req.URL.Path = singleJoiningSlash(target.Path, req.URL.Path)
		if targetQuery == "" || req.URL.RawQuery == "" {
			req.URL.RawQuery = targetQuery + req.URL.RawQuery
		} else {
			req.URL.RawQuery = targetQuery + "&" + req.URL.RawQuery
		}
		if _, ok := req.Header["User-Agent"]; !ok {
			// explicitly disable User-Agent so it's not set to default value
			req.Header.Set("User-Agent", "")
		}
	}
}

func (p *MultiHostReverseProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	p.Director = p.Rewrite(r)
	// Call upstream ServeHTTP
	p.ReverseProxy.ServeHTTP(w, r)
}
