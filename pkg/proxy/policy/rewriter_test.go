package policy

import (
	"context"
	log "github.com/owncloud/ocis-pkg/v2/log"
	"io"
	log_std "log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"
)

type directorFuncTestCase struct {
	name          string
	clientRequest *http.Request
	config        Route
	testFunc      func(t *testing.T, tc directorFuncTestCase)
}

func TestDirectorFunctions(t *testing.T) {
	var tests = []directorFuncTestCase{
		{"method_post", r("POST", "https://example.com/foo/", nil), Route{"/foo/", u("http://backend:8080"), false}, testBasicRewrite},
		{"method_put", r("POST", "https://example.com/foo/", nil), Route{"/foo/", u("http://backend:8080"), false}, testBasicRewrite},
		{"method_get", r("GET", "https://example.com/foo/", nil), Route{"/foo/", u("http://backend:8080"), false}, testBasicRewrite},
		{"with_query", r("GET", "https://example.com/foo?a=1&b=2", nil), Route{"/foo/", u("http://backend:8080"), false}, testBasicRewrite},
		{"root", r("GET", "https://example.com/", nil), Route{"/", u("http://backend:8080"), false}, testBasicRewrite},
		{"vhost_rewrite", r("GET", "https://example.com/", nil), Route{"/", u("http://backend:8080"), true}, testVHostRewrite},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			tc := tc
			tc.testFunc(t, tc)
		})
	}
}

func testBasicRewrite(t *testing.T, tc directorFuncTestCase) {
	director, err := newDirectorFn(&tc.config)
	if err != nil {
		t.Fatal(err)
	}

	origReq := tc.clientRequest
	gotReq := tc.clientRequest.Clone(context.Background())

	director(gotReq)

	var got, want = origReq.Method, gotReq.Method
	if origReq.Method != gotReq.Method {
		t.Errorf("Backend request should have method %v, got %v", origReq.Method, gotReq.Method)
	}

	got, want = gotReq.URL.Host, tc.config.Backend.Host
	if got != want {
		t.Errorf("Backend request should have url %v, got %v", origReq.URL.Host, gotReq.Host)
	}

	got, want = gotReq.URL.Path, origReq.URL.Path
	if got != want {
		t.Errorf("Backend request should have path %v, got %v", origReq.URL.Path, gotReq.URL.Path)
	}

	got, want = gotReq.URL.RawQuery, origReq.URL.RawQuery
	if got != want {
		t.Errorf("Backend request should have query %v, got %v", origReq.URL.RawQuery, gotReq.URL.RawQuery)
	}

}

func testVHostRewrite(t *testing.T, tc directorFuncTestCase) {
	director, err := newDirectorFn(&tc.config)
	if err != nil {
		t.Fatal(err)
	}

	origReq := tc.clientRequest
	gotReq := tc.clientRequest.Clone(context.Background())

	director(gotReq)
	var got, want = gotReq.Host, tc.config.Backend.Host
	if got != want {
		t.Errorf("Backend request should have host %v when apache-vhost option is set to true got %v", origReq.URL.RawQuery, gotReq.URL.RawQuery)
	}

}

type rewriterTestCase struct {
	in  *http.Request
	exp *url.URL
}

func TestRewriter(t *testing.T) {
	policies := []Policy{
		{
			Name: "default",
			Routes: []Route{
				{
					Endpoint: "/",
					Backend:  u("http://localhost:8080"),
				},
				{
					Endpoint: "/service1/",
					Backend:  u("http://localhost:1111"),
				},
				{
					Endpoint: "/service2/",
					Backend:  u("http://localhost:2222"),
				},
			},
		},
	}

	strategy := StaticPolicyStrategy(&StaticPolicyConfig{PolicyName: "default"})

	var rwTests = []rewriterTestCase{
		{
			in:  r("GET", "http://localhost:443/hello/world", nil),
			exp: u("http://localhost:8080/hello/world"),
		},
		{
			in:  r("GET", "http://localhost:443/service1/", nil),
			exp: u("http://localhost:1111/service1/"),
		},
		{
			in:  r("GET", "http://localhost:443/service1/foo/bar/baz", nil),
			exp: u("http://localhost:1111/service1/foo/bar/baz"),
		},
		{
			in:  r("GET", "http://localhost:443/service2/system.php", nil),
			exp: u("http://localhost:2222/service2/system.php"),
		},
		{
			in:  r("GET", "http://localhost:443/service2/system.php?q1=a&q2=b", nil),
			exp: u("http://localhost:2222/service2/system.php?q1=a&q2=b"),
		},
	}

	prw, err := NewPolicyRewriter(policies, strategy, log.NewLogger())
	if err != nil {
		t.Error(err)
	}

	for k := range rwTests {
		t.Run(strconv.Itoa(k), func(t *testing.T) {
			t.Parallel()
			tc := rwTests[k]
			req := tc.in.Clone(context.Background())
			prw(req)(req)

			if req.URL.String() != tc.exp.String() {
				t.Errorf("Expected rewriten url to be %v, got %v", tc.exp.String(), tc.exp.String())
			}
		})

	}
}

func r(m string, target string, body io.Reader) *http.Request {
	return httptest.NewRequest(m, target, body)
}

func u(strURL string) *url.URL {
	pURL, err := url.Parse(strURL)
	if err != nil {
		log_std.Fatalf("Error parsing url: %v", err)
	}

	return pURL
}
