package policy

import (
	"context"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

type testCase struct {
	name          string
	clientRequest *http.Request
	config        Route
	testFunc      func(t *testing.T, tc testCase)
}

func TestNewDirectorFn(t *testing.T) {
	var tests = []testCase{
		{"method_post", r("POST", "https://example.com/foo/", nil), Route{"/foo/", u("http://backend:8080"), false}, testBasicRewrite},
		{"method_put", r("POST", "https://example.com/foo/", nil), Route{"/foo/", u("http://backend:8080"), false}, testBasicRewrite},
		{"method_get", r("GET", "https://example.com/foo/", nil), Route{"/foo/", u("http://backend:8080"), false}, testBasicRewrite},
		{"with_query", r("GET", "https://example.com/foo?a=1&b=2", nil), Route{"/foo/", u("http://backend:8080"), false}, testBasicRewrite},
		{"root", r("GET", "https://example.com/", nil), Route{"/", u("http://backend:8080"), false}, testBasicRewrite},
		{"vhost_rewrite", r("GET", "https://example.com/", nil), Route{"/", u("http://backend:8080"), true}, testVHostRewrite},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tc := tc
			tc.testFunc(t, tc)
		})
	}

}

func testBasicRewrite(t *testing.T, tc testCase) {
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

func testVHostRewrite(t *testing.T, tc testCase) {
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

func r(m string, target string, body io.Reader) *http.Request {
	return httptest.NewRequest(m, target, body)
}

func u(strUrl string) *url.URL {
	pUrl, err := url.Parse(strUrl)
	if err != nil {
		log.Fatalf("Error parsing url: %v", err)
	}

	return pUrl
}
