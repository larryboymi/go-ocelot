package routes

import (
	"encoding/json"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"

	"github.com/larryboymi/go-ocelot/cache"
	"github.com/larryboymi/go-ocelot/poller"
	"github.com/larryboymi/go-ocelot/types"
)

// Synchronizer
type Synchronizer struct {
	interval    time.Duration
	cache       cache.Cache
	routePoller poller.Poller
	routes      map[string]types.Route
}

// Routes accessor
func (s *Synchronizer) Routes() map[string]types.Route {
	return s.routes
}

// Handler for serving requests
func (s *Synchronizer) Handler(w http.ResponseWriter, req *http.Request) {
	log.Printf("Routing for %s", req.URL.Path)
	w.Header().Set("powered-by", "go-ocelot")
	if req.TLS != nil {
		req.Header.Add("x-forwarded-proto", "https")
	} else {
		req.Header.Add("x-forwarded-proto", "http")
	}
	s.routes["ecgo"].Proxy.ServeHTTP(w, req)
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

	s.routes = newTable
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
	}
}

// Start causes the collector to begin polling docker
func (s *Synchronizer) Start() {
	go func() {
		s.syncRoutes()
		for {
			err := s.cache.Subscribe("go-ocelot", s.syncRoutes)
			log.Printf("Subscription to updates lost, retrying in 10 seconds: %v", err)
			time.Sleep(10 * time.Second)
		}
	}()

	go func() {
		s.updateRoutes()

		for range time.Tick(s.interval * time.Second) {
			s.updateRoutes()
		}
	}()
}

// New returns a new instance of the collector
func New(interval int, redis string) Synchronizer {
	return Synchronizer{
		interval:    time.Duration(interval),
		cache:       cache.New(redis),
		routePoller: poller.New(),
	}
}
