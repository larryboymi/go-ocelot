package routes

import (
	"encoding/json"
	"log"
	"time"

	"github.com/larryboymi/go-ocelot/cache"
	"github.com/larryboymi/go-ocelot/poller"
)

// RouteMaintainer
type RouteMaintainer struct {
	interval    time.Duration
	cache       cache.Cache
	routePoller poller.Poller
}

func (r *RouteMaintainer) updateRoutes() {
	log.Print("Updating routes")
	routes := r.routePoller.Load()

	json, err := json.Marshal(routes)

	if err != nil {
		log.Printf("Error marshalling routes to JSON: %v", err)
		return
	}

	if cacheError := r.cache.Set("routes", string(json)); cacheError != nil {
		log.Printf("Error storing routes in cache: %v", cacheError)
	} else {
		log.Printf("Stored routes in cache: %s", string(json))
	}
}

// Start causes the collector to begin polling docker
func (r *RouteMaintainer) Start() {
	r.updateRoutes()

	for range time.Tick(r.interval * time.Second) {
		r.updateRoutes()
	}
}

// New returns a new instance of the collector
func New(interval int, redis string) RouteMaintainer {
	return RouteMaintainer{
		interval:    time.Duration(interval),
		cache:       cache.New(redis),
		routePoller: poller.New(),
	}
}
