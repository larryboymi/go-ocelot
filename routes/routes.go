/*
Package routes helps to maintain a set of routes retrieved from the Docker sock
*/
package routes

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"regexp"
	"sync"
	"time"

	"github.com/larryboymi/go-ocelot/cache"
	"github.com/larryboymi/go-ocelot/poller"
	"github.com/larryboymi/go-ocelot/types"
)

// Synchronizer is the type that maintains the route set from Docker sock.
type Synchronizer struct {
	interval    time.Duration
	cache       cache.Cache
	routePoller poller.Poller
	routes      *SafeRoutes
}

// SafeRoutes helps to ensure only one thread is accessing routes via mutex
type SafeRoutes struct {
	routes map[string]types.Route
	mux    sync.Mutex
}

// Routes accessor
func (s *Synchronizer) Routes() map[string]types.Route {
	s.routes.mux.Lock()
	defer s.routes.mux.Unlock()
	return s.routes.routes
}

func (s *Synchronizer) syncRoutesFromRedis() {
	var routes []types.Route
	routesJSON, getErr := s.cache.GetAll("routes")
	if getErr != nil {
		log.Printf("Error loading routes: %v", getErr)
		return
	}
	for _, routeStr := range routesJSON {
		var route types.Route
		err := json.Unmarshal([]byte(routeStr), &route)
		if err != nil {
			log.Print("Error syncing routes", err)
			return
		}
		routes = append(routes, route)
	}

	s.updateRoutingTable(routes...)
	log.Printf("Updated routes successfully")
}

// UpdateRoutes is an atomic operation to update the routing table
func (s *Synchronizer) updateRoutingTable(routes ...types.Route) {
	s.routes.mux.Lock()
	defer s.routes.mux.Unlock()
	for _, route := range routes {
		s.routes.routes[route.ID] = s.augmentRoute(route)
	}
}

func (s *Synchronizer) augmentRoute(route types.Route) types.Route {
	var err error
	parsedURL, _ := url.Parse(route.TargetURL)
	route.Regexp, err = regexp.Compile(route.IncomingMatch)
	if err != nil {
		log.Printf("Error compiling regexp for %s: %v", route.ID, err)
	}
	route.Proxy = httputil.NewSingleHostReverseProxy(parsedURL)
	return route
}

// DeleteRoute updates a route both in redis and in memory
func (s *Synchronizer) DeleteRoute(id string) (int, error) {
	s.routes.mux.Lock()
	defer s.routes.mux.Unlock()
	if _, ok := s.routes.routes[id]; ok {
		delete(s.routes.routes, id)
	} else {
		return http.StatusNotFound, fmt.Errorf("Route not found for %s", id)
	}
	if cacheError := s.cache.DeleteField("routes", id); cacheError != nil {
		log.Printf("Error removing route from cache: %v", cacheError)
		return http.StatusInternalServerError, fmt.Errorf("Error removing route from cache: %v", cacheError)
	}
	log.Printf("Removed route from cache: %s", id)
	return http.StatusNoContent, nil
}

// UpdateRoute updates a route both in redis and in memory
func (s *Synchronizer) UpdateRoute(route types.Route) {
	s.updateRoutingTable(route)
	json, err := json.Marshal(route)
	if err != nil {
		log.Printf("Error marshalling route to JSON: %v", err)
		return
	}
	if cacheError := s.cache.SetField("routes", route.ID, string(json)); cacheError != nil {
		log.Printf("Error storing route in cache: %v", cacheError)
	} else {
		log.Printf("Stored route in cache: %s", string(json))
	}
}

func (s *Synchronizer) updateRoutesFromDocker() {
	log.Print("Updating routes")
	dockerRoutes := s.routePoller.Load()
	for _, dockerRoute := range dockerRoutes {
		// If route in memory doesn't exist, update redis, then add to memory
		if s.routes.routes[dockerRoute.ID].ID == "" {
			s.UpdateRoute(dockerRoute)
		}
	}
}

// Start causes the Synchronizer to begin polling docker
func (s *Synchronizer) Start() {
	// This thread will continually get the routes from the cache and make them available
	go func() {
		s.syncRoutesFromRedis()
		for {
			err := s.cache.Subscribe("go-ocelot", s.syncRoutesFromRedis)
			log.Printf("Subscription to updates lost, retrying in 10 seconds: %v", err)
			time.Sleep(10 * time.Second)
		}
	}()

	// This thread will continually get the routes from the Docker sock and set them in the cache
	go func() {
		s.updateRoutesFromDocker()

		for range time.Tick(s.interval * time.Second) {
			s.updateRoutesFromDocker()
		}
	}()
}

// New returns a new instance of the synchronizer
func New(interval int, redis string) Synchronizer {
	return Synchronizer{
		interval:    time.Duration(interval),
		cache:       cache.New(redis),
		routePoller: poller.New(),
		routes:      &SafeRoutes{routes: make(map[string]types.Route), mux: sync.Mutex{}},
	}
}
