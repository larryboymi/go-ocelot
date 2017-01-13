package proxy

import (
	"net/http"
	"net/http/httputil"
	"net/url"
)

// Proxy type
type Proxy struct {
	target *url.URL
	proxy  *httputil.ReverseProxy
}

// New Proxy
func New(target string) *Proxy {
	url, _ := url.Parse(target)
	// you should handle error on parsing
	return &Proxy{target: url, proxy: httputil.NewSingleHostReverseProxy(url)}
}

// Handler for serving requests
func (p *Proxy) Handler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("powered-by", "go-ocelot")
	if r.TLS != nil {
		r.Header.Add("x-forwarded-proto", "https")
	} else {
		r.Header.Add("x-forwarded-proto", "http")
	}
	p.proxy.ServeHTTP(w, r)
}
