package proxy

import (
	"github.com/owncloud/ocis-pkg/v2/log"
	ocisoidc "github.com/owncloud/ocis-pkg/v2/oidc"
	"github.com/owncloud/ocis-proxy/pkg/proxy/policy"
	"net/http"
	"net/http/httputil"
	"strings"
)

// MultiHostReverseProxy extends httputil to support multiple hosts with different policies.
type MultiHostReverseProxy struct {
	httputil.ReverseProxy
	Directors policy.Directors
	Strategy  policy.Strategy
	logger    log.Logger
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

	directors, err := policy.NewDirectors(options.Config.Policies)
	if err != nil {
		l.Fatal().Err(err).Msgf("Could not load policies")
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

	reverseProxy := &MultiHostReverseProxy{
		Directors: directors,
		Strategy:  policy.Migration(),
		logger:    l,
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
	var pol policy.Name
	var userID string
	claims := ocisoidc.FromContext(r.Context())

	if claims != nil {
		userID = claims.PreferredUsername
	}

	pol = p.Strategy.Policy(r.Context(), userID)
	p.routeWithPolicy(pol, r)

	// Call upstream ServeHTTP
	p.ReverseProxy.ServeHTTP(w, r)
}

func (p *MultiHostReverseProxy) routeWithPolicy(policy policy.Name, r *http.Request) {
	var hit bool

	if _, ok := p.Directors[policy]; !ok {
		p.logger.
			Error().
			Msgf("policy %v is not configured", policy)
	}

	for k := range p.Directors[policy] {
		ep := string(k)
		if strings.HasPrefix(r.URL.Path, ep) && ep != "/" {
			p.Director = p.Directors[policy][k]
			hit = true
			p.logger.
				Debug().
				Interface("policy", policy).
				Interface("endpoint", ep).
				Interface("path", r.URL.Path).
				Msg("director found")
		}
	}

	// override default director with root. If any
	if !hit && p.Directors[policy]["/"] != nil {
		p.Director = p.Directors[policy]["/"]
	}

	// Call upstream ServeHTTP
	p.ReverseProxy.ServeHTTP(w, r)
}

func defaultPolicies() []config.Policy {
	return []config.Policy{
		config.Policy{
			Name: "reva",
			Routes: []config.Route{
				config.Route{
					Endpoint: "/",
					Backend:  "http://localhost:9100",
				},
				config.Route{
					Endpoint: "/.well-known/",
					Backend:  "http://localhost:9130",
				},
				config.Route{
					Endpoint: "/konnect/",
					Backend:  "http://localhost:9130",
				},
				config.Route{
					Endpoint: "/signin/",
					Backend:  "http://localhost:9130",
				},
				config.Route{
					Endpoint: "/ocs/",
					Backend:  "http://localhost:9140",
				},
				config.Route{
					Endpoint: "/remote.php/",
					Backend:  "http://localhost:9140",
				},
				config.Route{
					Endpoint: "/dav/",
					Backend:  "http://localhost:9140",
				},
				config.Route{
					Endpoint: "/webdav/",
					Backend:  "http://localhost:9140",
				},
			},
		},
		config.Policy{
			Name: "oc10",
			Routes: []config.Route{
				config.Route{
					Endpoint: "/",
					Backend:  "http://localhost:9100",
				},
				config.Route{
					Endpoint: "/.well-known/",
					Backend:  "http://localhost:9130",
				},
				config.Route{
					Endpoint: "/konnect/",
					Backend:  "http://localhost:9130",
				},
				config.Route{
					Endpoint: "/signin/",
					Backend:  "http://localhost:9130",
				},
				config.Route{
					Endpoint:    "/ocs/",
					Backend:     "https://demo.owncloud.com",
					ApacheVHost: true,
				},
				config.Route{
					Endpoint:    "/remote.php/",
					Backend:     "https://demo.owncloud.com",
					ApacheVHost: true,
				},
				config.Route{
					Endpoint:    "/dav/",
					Backend:     "https://demo.owncloud.com",
					ApacheVHost: true,
				},
				config.Route{
					Endpoint:    "/webdav/",
					Backend:     "https://demo.owncloud.com",
					ApacheVHost: true,
				},
			},
		},
	}
}
