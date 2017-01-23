package routes

import (
	"fmt"
	"log"
	"strings"

	"github.com/ocelotconsulting/go-ocelot/types"
)

// Handler for serving requests
func findRoute(key string, routes map[string]types.Route) *types.Route {
	for _, route := range routes {
		if route.ProxiedURL == key {
			return &route
		}
	}
	return nil
}

func findRouteByPath(url string, pathDepth int, routes map[string]types.Route) *types.Route {
	if pathDepth == 0 {
		return nil
	}
	key := strings.Join(strings.SplitN(url, "/", pathDepth), "/")
	log.Printf("Searching for route with key %s", key)
	if route := findRoute(key, routes); route != nil {
		return route
	} else if route := findRoute(fmt.Sprintf("www.%s", key), routes); route != nil {
		return route
	} else {
		return findRouteByPath(key, pathDepth-1, routes)
	}
}

//ResolveRoute helps the proxy find a route for the incoming request
func ResolveRoute(url, host string, routes map[string]types.Route) *types.Route {
	url = strings.Split(url, "?")[0]
	if closestRoute := findRouteByPath(fmt.Sprintf("%s%s", host, url), 4, routes); closestRoute != nil {
		return closestRoute
	}
	return nil
}
