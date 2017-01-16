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
func (p *Proxy) Handler(w http.ResponseWriter, req *http.Request) {
	log.Printf("Routing for %s", req.URL.Path)
	w.Header().Set("powered-by", "go-ocelot")
	if req.TLS != nil {
		req.Header.Add("x-forwarded-proto", "https")
	} else {
		req.Header.Add("x-forwarded-proto", "http")
	}
	p.sync.Routes()["ecgo"].Proxy.ServeHTTP(w, req)
}

// New returns a new instance of the proxy
func New(sync *routes.Synchronizer) Proxy {
	return Proxy{
		sync: sync,
	}
}
