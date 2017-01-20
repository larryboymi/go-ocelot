package routes

import (
	"fmt"
	"log"
	"strings"

	"github.com/larryboymi/go-ocelot/types"
)

// Handler for serving requests
func (s *Synchronizer) findRoute(key string) *types.Route {
	for _, route := range s.Routes() {
		if route.ProxiedURL == key {
			return &route
		}
	}
	return nil
}

func (s *Synchronizer) findRouteByPath(url string, pathDepth int) *types.Route {
	if pathDepth == 0 {
		return nil
	}
	key := strings.Join(strings.SplitN(url, "/", pathDepth), "/")
	log.Printf("Searching for route with key %s", key)
	if route := s.findRoute(key); route != nil {
		return route
	} else if route := s.findRoute(fmt.Sprintf("www.%s", key)); route != nil {
		return route
	} else {
		return s.findRouteByPath(key, pathDepth-1)
	}
}

//ResolveRoute helps the proxy find a route for the incoming request
func (s *Synchronizer) ResolveRoute(url, host string) *types.Route {
	url = strings.Split(url, "?")[0]
	if closestRoute := s.findRouteByPath(fmt.Sprintf("%s%s", host, url), 4); closestRoute != nil {
		return closestRoute
	}
	return nil
}
