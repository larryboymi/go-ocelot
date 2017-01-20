package proxy

import (
	"log"
	"net/http"
	"net/http/httputil"

	"github.com/larryboymi/go-ocelot/routes"
)

// Proxy is the type that handles routing requests to the correct proxied item
type Proxy struct {
	sync  *routes.Synchronizer
	proxy *httputil.ReverseProxy
}

// Handler for serving requests
func (p *Proxy) Handler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Proxy handler trying to route %s with path %s", r.Host, r.URL.Path)
	pathToMatch := ""
	if r.URL.Path != "/" {
		pathToMatch = r.URL.Path
	}
	if route := p.sync.ResolveRoute(pathToMatch, r.Host); route != nil {
		p.proxy.ServeHTTP(w, r)
		return
	}
	// no pattern matched; send 404 response
	http.NotFound(w, r)
}

// New returns a new instance of the proxy
func New(sync *routes.Synchronizer) Proxy {
	return Proxy{
		sync:  sync,
		proxy: sync.NewReverseProxyHTTP(),
	}
}
