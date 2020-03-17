package policy

import (
	"github.com/owncloud/ocis-pkg/v2/log"
	ocisoidc "github.com/owncloud/ocis-pkg/v2/oidc"
	"net/http"
	"strings"
)

type (
	Name      string
	Endpoint  string
	Directors map[Name]map[Endpoint]func(req *http.Request)
)

func NewPolicyRewriter(policies []Policy, policyStrategy Strategy, logger log.Logger) (func(r *http.Request) func(r *http.Request), error) {
	var err error
	var directors Directors
	var l log.Logger
	directors, err = loadDirectors(policies)
	l = logger

	if err != nil {
		return nil, err
	}

	strategy := policyStrategy

	return func(r *http.Request) func(r *http.Request) {
		var pol Name
		var userID string
		var hit bool

		claims := ocisoidc.FromContext(r.Context())

		if claims != nil {
			userID = claims.PreferredUsername
		}

		pol = strategy.Policy(r.Context(), userID)

		if _, ok := directors[pol]; !ok {
			l.Error().Msgf("policy %v is not configured", pol)
		}

		for k := range directors[pol] {
			ep := string(k)
			if strings.HasPrefix(r.URL.Path, ep) && ep != "/" {
				hit = true
				l.Debug().
					Interface("policy", pol).
					Interface("endpoint", ep).
					Interface("path", r.URL.Path).
					Msg("director found")

				return directors[pol][k]
			}
		}

		// override default director with root. If any
		if !hit && directors[pol]["/"] != nil {
			return directors[pol]["/"]
		}

		l.Error().Msgf("No director found and no root-route configured")
		return func(r *http.Request) {

		}

	}, nil
}

func loadDirectors(policies []Policy) (Directors, error) {
	directors := make(Directors)
	for _, policy := range policies {
		for _, route := range policy.Routes {
			if directors[policy.Name] == nil {
				directors[policy.Name] = make(map[Endpoint]func(req *http.Request))
			}

			dir, err := newDirectorFn(&route)
			if err != nil {
				return nil, err
			}

			directors[policy.Name][route.Endpoint] = dir
		}
	}

	return directors, nil
}

func newDirectorFn(rt *Route) (func(req *http.Request), error) {
	route := rt
	target := route.Backend
	targetQuery := target.RawQuery

	return func(req *http.Request) {
		req.URL.Scheme = target.Scheme
		req.URL.Host = target.Host
		if route.ApacheVHost {
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
	}, nil
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
