/*
Package routes helps to maintain a set of routes retrieved from the Docker sock
*/
package routes

import (
	"encoding/json"
	"log"
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

func (s *Synchronizer) syncRoutes() {
	var routes []types.Route
	routesJSON, getErr := s.cache.Get("routes")

	if getErr != nil {
		log.Printf("Error loading routes: %v", getErr)
		return
	}

	err := json.Unmarshal(routesJSON, &routes)

	if err != nil {
		log.Print("Error syncing routes", err)
		return
	}

	s.UpdateRoutes(routes)
	log.Printf("Updated routes successfully")
}

// UpdateRoutes is an atomic operation to update the routing table
func (s *Synchronizer) UpdateRoutes(routes []types.Route) {
	for _, route := range routes {
		s.UpdateRoute(route)
	}
}

// UpdateRoute updates a route both in redis and in memory
func (s *Synchronizer) UpdateRoute(route types.Route) {
	var err error
	s.routes.mux.Lock()
	defer s.routes.mux.Unlock()
	url, _ := url.Parse(route.TargetURL)
	route.Regexp, err = regexp.Compile(route.IncomingMatch)
	if err != nil {
		log.Printf("Error compiling regexp for %s: %v", route.ID, err)
		err = nil
	}
	route.Proxy = httputil.NewSingleHostReverseProxy(url)
	json, err := json.Marshal(route)
	if err != nil {
		log.Printf("Error marshalling route to JSON: %v", err)
		return
	}
	if cacheError := s.cache.SetField("routes", route.ID, string(json)); cacheError != nil {
		log.Printf("Error storing route in cache: %v", cacheError)
	} else {
		log.Printf("Stored route in cache: %s", string(json))
		s.routes.routes[route.ID] = route
	}
}

func (s *Synchronizer) updateRoutes() {
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
		s.syncRoutes()
		for {
			err := s.cache.Subscribe("go-ocelot", s.syncRoutes)
			log.Printf("Subscription to updates lost, retrying in 10 seconds: %v", err)
			time.Sleep(10 * time.Second)
		}
	}()

	// This thread will continually get the routes from the Docker sock and set them in the cache
	go func() {
		s.updateRoutes()

		for range time.Tick(s.interval * time.Second) {
			s.updateRoutes()
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
