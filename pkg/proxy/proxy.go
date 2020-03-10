package proxy

import (
	"context"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/owncloud/ocis-pkg/v2/log"
	"github.com/owncloud/ocis-proxy/pkg/config"

	"go.opencensus.io/plugin/ochttp/propagation/tracecontext"
	"go.opencensus.io/trace"
)

// Directors map strings to httputil ReverseProxy Director
type Directors map[string]func(req *http.Request)

// MultiHostReverseProxy extends httputil to support multiple hosts with diffent policies
type MultiHostReverseProxy struct {
	httputil.ReverseProxy
	DirMap     map[string]Directors
	propagator tracecontext.HTTPFormat
	logger     log.Logger
	config     config.Config
}

// NewMultiHostReverseProxy undocummented
func NewMultiHostReverseProxy(opts ...Option) *MultiHostReverseProxy {
	options := newOptions(opts...)

	reverseProxy := &MultiHostReverseProxy{
		DirMap: make(map[string]Directors),
		logger: options.Logger,
		config: *options.Config,
	}

	for _, policy := range options.Config.Policies {
		for _, route := range policy.Routes {
			uri, err := url.Parse(route.Backend)
			if err != nil {
				reverseProxy.logger.
					Fatal().
					Err(err).
					Msgf("malformed url: %v", route.Backend)
			}

			reverseProxy.logger.
				Debug().
				Interface("route", route).
				Msg("adding route")

			reverseProxy.AddHost(policy.Name, uri, route)
		}
	}

	return reverseProxy
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
	if p.DirMap[policy] == nil {
		p.DirMap[policy] = make(map[string]func(req *http.Request))
	}

	p.DirMap[policy][rt.Endpoint] = func(req *http.Request) {
		req.URL.Scheme = target.Scheme
		req.URL.Host = target.Host
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

	// TODO need to fetch from the accounts service
	// ============
	policy := "reva"
	// ============

	ctx := context.Background()
	var span *trace.Span

	// Start a root span
	if p.config.Tracing.Enabled {
		ctx, span = trace.StartSpan(context.Background(), r.URL.String())
		defer span.End()
		p.propagator.SpanContextToRequest(span.SpanContext(), r)
	}

	if _, ok := p.DirMap[policy]; !ok {
		p.logger.Fatal().Msgf("policy %v is not configured", policy)
	}

	// override default director with root. TODO this is a design flaw, as we need a catch-all director
	if p.DirMap[policy]["/"] != nil {
		p.Director = p.DirMap[policy]["/"]
	}

	for k := range p.DirMap[policy] {
		if strings.HasPrefix(r.URL.Path, k) && k != "/" {
			p.Director = p.DirMap[policy][k]
			p.logger.Debug().Str("policy", policy).Str("prefix", k).Str("path", r.URL.Path).Msg("director found")
		}
	}

	// Call upstream ServeHTTP
	p.ReverseProxy.ServeHTTP(w, r.WithContext(ctx))
}
