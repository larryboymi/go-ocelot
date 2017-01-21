package proxy

import (
	"log"
	"net/http"

	"github.com/larryboymi/go-ocelot/proxy/reverse"
	"github.com/larryboymi/go-ocelot/routes"
)

//New returns a handler that will proxy incoming requests
func New(repo routes.Repository) http.Handler {
	proxy := reverse.New(repo)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Proxy handler trying to route %s with path %s", r.Host, r.URL.Path)
		pathToMatch := ""
		if r.URL.Path != "/" {
			pathToMatch = r.URL.Path
		}
		if route := routes.ResolveRoute(pathToMatch, r.Host, repo.Routes()); route != nil {
			proxy.ServeHTTP(w, r)
			return
		}
		// no pattern matched; send 404 response
		http.NotFound(w, r)
	})
}
