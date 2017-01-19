package proxy

import (
	"fmt"
	"strings"

	"github.com/larryboymi/go-ocelot/types"
)

// Handler for serving requests
func (p *Proxy) findRoute(key string) *types.Route {
	if route, ok := p.sync.Routes()[key]; ok {
		return &route
	}
	return nil
}

func (p *Proxy) findRouteByPath(url string, pathDepth int) *types.Route {
	if pathDepth == 0 {
		return nil
	}
	key := strings.Join(strings.SplitN(url, "/", pathDepth), "/")
	if route := p.findRoute(key); route != nil {
		return route
	} else if route := p.findRoute(fmt.Sprintf("www.%s", key)); route != nil {
		return route
	} else {
		return p.findRouteByPath(key, pathDepth-1)
	}
}

//ResolveRoute helps the proxy find a route for the incoming request
func (p *Proxy) ResolveRoute(url, host string) *types.Route {
	url = strings.Split(url, "?")[0]
	if closestRoute := p.findRouteByPath(fmt.Sprintf("%s%s", host, url), 4); closestRoute != nil {
		return closestRoute
	}
	return nil
}