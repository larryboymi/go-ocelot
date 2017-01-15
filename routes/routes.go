package routes

import (
	"encoding/json"
	"log"
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

func (r *Synchronizer) syncRoutes() {
	var routes []types.Route
	routesJSON, getErr := r.cache.Get("routes")

	if getErr != nil {
		log.Printf("Error loading routes: %v", getErr)
		return
	}

	err := json.Unmarshal(routesJSON, &routes)

	if err != nil {
		log.Print("Error syncing routes", err)
		return
	}

	r.UpdateRoutes(routes)
	log.Printf("Updated routes successfully")
}

// UpdateTable is an atomic operation to update the routing table
func (r *Synchronizer) UpdateRoutes(routes []types.Route) {
	newTable := make(map[string]types.Route)

	for _, s := range routes {
		log.Printf("Saving route %s", s.ID)
		newTable[s.ID] = s
	}

	r.routes = newTable
}

func (r *Synchronizer) updateRoutes() {
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
func (r *Synchronizer) Start() {
	go func() {
		r.syncRoutes()
		for {
			err := r.cache.Subscribe("go-ocelot", r.syncRoutes)
			log.Printf("Subscription to updates lost, retrying in 10 seconds: %v", err)
			time.Sleep(10 * time.Second)
		}
	}()

	r.updateRoutes()

	for range time.Tick(r.interval * time.Second) {
		r.updateRoutes()
	}
}

// New returns a new instance of the collector
func New(interval int, redis string) Synchronizer {
	return Synchronizer{
		interval:    time.Duration(interval),
		cache:       cache.New(redis),
		routePoller: poller.New(),
	}
}
