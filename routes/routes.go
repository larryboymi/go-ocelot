/*
Package routes helps to maintain a set of routes retrieved from the Docker sock
*/
package routes

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/ocelotconsulting/go-ocelot/cache"
	"github.com/ocelotconsulting/go-ocelot/poller"
	"github.com/ocelotconsulting/go-ocelot/types"
)

// Repository contains the routes for the proxy
type Repository interface {
	Routes() map[string]types.Route
	DeleteRoute(id string) (int, error)
	UpdateRoute(route types.Route)
	Start()
}

type routeWrapper struct {
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
func (r *routeWrapper) Routes() map[string]types.Route {
	r.routes.mux.Lock()
	defer r.routes.mux.Unlock()
	return r.routes.routes
}

func (r *routeWrapper) syncRoutesFromRedis() {
	var routes []types.Route
	routesJSON, getErr := r.cache.GetAll("routes")
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

	r.updateRoutingTable(routes...)
	log.Printf("Updated routes successfully")
}

// UpdateRoutes is an atomic operation to update the routing table
func (r *routeWrapper) updateRoutingTable(routes ...types.Route) {
	r.routes.mux.Lock()
	defer r.routes.mux.Unlock()
	for _, route := range routes {
		r.routes.routes[route.ID] = route
	}
}

// DeleteRoute updates a route both in redis and in memory
func (r *routeWrapper) DeleteRoute(id string) (int, error) {
	r.routes.mux.Lock()
	defer r.routes.mux.Unlock()
	if _, ok := r.routes.routes[id]; ok {
		delete(r.routes.routes, id)
	} else {
		return http.StatusNotFound, fmt.Errorf("Route not found for %s", id)
	}
	if cacheError := r.cache.DeleteField("routes", id); cacheError != nil {
		log.Printf("Error removing route from cache: %v", cacheError)
		return http.StatusInternalServerError, fmt.Errorf("Error removing route from cache: %v", cacheError)
	}
	log.Printf("Removed route from cache: %s", id)
	return http.StatusNoContent, nil
}

// UpdateRoute updates a route both in redis and in memory
func (r *routeWrapper) UpdateRoute(route types.Route) {
	r.updateRoutingTable(route)
	json, err := json.Marshal(route)
	if err != nil {
		log.Printf("Error marshalling route to JSON: %v", err)
		return
	}
	if cacheError := r.cache.SetField("routes", route.ID, string(json)); cacheError != nil {
		log.Printf("Error storing route in cache: %v", cacheError)
	} else {
		log.Printf("Stored route in cache: %s", string(json))
	}
}

func (r *routeWrapper) updateRoutesFromDocker() {
	log.Print("Updating routes")
	dockerRoutes := r.routePoller.Load()
	for _, dockerRoute := range dockerRoutes {
		// If route in memory doesn't exist, update redis, then add to memory
		if r.routes.routes[dockerRoute.ID].ID == "" {
			r.UpdateRoute(dockerRoute)
		}
	}
}

// Start causes the Synchronizer to begin polling docker
func (r *routeWrapper) Start() {
	// This thread will continually get the routes from the cache and make them available
	go func() {
		r.syncRoutesFromRedis()
		for {
			err := r.cache.Subscribe("go-ocelot", r.syncRoutesFromRedis)
			log.Printf("Subscription to updates lost, retrying in 10 seconds: %v", err)
			time.Sleep(10 * time.Second)
		}
	}()

	// This thread will continually get the routes from the Docker sock and set them in the cache
	go func() {
		r.updateRoutesFromDocker()

		for range time.Tick(r.interval * time.Second) {
			r.updateRoutesFromDocker()
		}
	}()
}

// New returns a new instance of the synchronizer
func New(interval int, redis string) Repository {
	return Repository(&routeWrapper{
		interval:    time.Duration(interval),
		cache:       cache.New(redis),
		routePoller: poller.New(),
		routes:      &SafeRoutes{routes: make(map[string]types.Route), mux: sync.Mutex{}},
	})
}
