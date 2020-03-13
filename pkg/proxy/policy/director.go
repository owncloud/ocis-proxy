package policy

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

type (
	Name      string
	Endpoint  string
	Directors map[Name]map[Endpoint]func(req *http.Request)
)

func NewDirectors(policies []Policy) (Directors, error) {
	directors := make(Directors)
	for _, policy := range policies {
		for _, route := range policy.Routes {
			uri, err := url.Parse(route.Backend)
			if err != nil {
				return nil, fmt.Errorf("malformed url: %v: %w", route.Backend, err)
			}

			if directors[policy.Name] == nil {
				directors[policy.Name] = make(map[Endpoint]func(req *http.Request))
			}

			directors[policy.Name][route.Endpoint] = newDirectorFn(uri, route)
		}
	}

	return directors, nil
}

func newDirectorFn(target *url.URL, rt Route) func(req *http.Request) {
	targetQuery := target.RawQuery

	return func(req *http.Request) {
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
