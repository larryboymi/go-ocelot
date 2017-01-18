package proxy

import (
	"log"
	"net/http"

	"github.com/larryboymi/go-ocelot/routes"
)

// Proxy is the type that handles routing requests to the correct proxied item
type Proxy struct {
	sync *routes.Synchronizer
}

// Handler for serving requests
func (p *Proxy) Handler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Proxy handler trying to route %s with path %s", r.URL.Host, r.URL.Path)
	for _, route := range p.sync.Routes() {
		if route.Regexp.MatchString(r.URL.Path) {
			route.Proxy.ServeHTTP(w, r)
			return
		}
	}
	// no pattern matched; send 404 response
	http.NotFound(w, r)
}

// New returns a new instance of the proxy
func New(sync *routes.Synchronizer) Proxy {
	return Proxy{
		sync: sync,
	}
}
