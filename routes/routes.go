/*
Package routes helps to maintain a set of routes retrieved from the Docker sock
*/
package routes

import (
	"encoding/json"
	"log"
	"net/http/httputil"
	"net/url"
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
	newTable := make(map[string]types.Route)

	for _, s := range routes {
		log.Printf("Saving route %s", s.ID)
		url, _ := url.Parse(s.TargetURL)
		s.Proxy = httputil.NewSingleHostReverseProxy(url)
		newTable[s.ID] = s
	}

	s.routes.mux.Lock()
	s.routes.routes = newTable
	s.routes.mux.Unlock()
}

func (s *Synchronizer) updateRoutes() {
	log.Print("Updating routes")
	routes := s.routePoller.Load()

	json, err := json.Marshal(routes)

	if err != nil {
		log.Printf("Error marshalling routes to JSON: %v", err)
		return
	}
	if cacheError := s.cache.Set("routes", string(json)); cacheError != nil {
		log.Printf("Error storing routes in cache: %v", cacheError)
	} else {
		log.Printf("Stored routes in cache: %s", string(json))
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
		routes:      &SafeRoutes{mux: sync.Mutex{}},
	}
}
