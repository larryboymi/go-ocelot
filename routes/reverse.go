package routes

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"strings"
)

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

// NewSingleHostReverseProxy returns a new ReverseProxy that routes
// URLs to the scheme, host, and base path provided in target. If the
// target's path is "/base" and the incoming request was for "/dir",
// the target request will be for /base/dir.
// NewSingleHostReverseProxy does not rewrite the Host header.
// To rewrite Host headers, use ReverseProxy directly with a custom
// Director policy.
func (s *Synchronizer) NewReverseProxyHTTP() *httputil.ReverseProxy {
	director := func(req *http.Request) {
		req.URL.Scheme = "http" // terminate ssl here

		pathToMatch := ""
		if req.URL.Path != "/" {
			pathToMatch = req.URL.Path
		}
		if route := s.ResolveRoute(pathToMatch, req.Host); route != nil {
			req.URL.Host = fmt.Sprintf("%s:%d", route.ID, route.TargetPort)
			req.URL.Path = singleJoiningSlash("", req.URL.Path)
			if _, ok := req.Header["User-Agent"]; !ok {
				// explicitly disable User-Agent so it's not set to default value
				req.Header.Set("User-Agent", "")
			}
		}
	}
	return &httputil.ReverseProxy{Director: director}
}
