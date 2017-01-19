package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/larryboymi/go-ocelot/routes"
	"github.com/larryboymi/go-ocelot/types"
)

// Service is the interface for stored services
// type Handler interface {
// 	GetRoutes(w http.ResponseWriter, req *http.Request) ([]types.Route, error)
// }

// API is the type that handles routing api requests for go-ocelot
type API struct {
	sync *routes.Synchronizer
}

func (a API) echo(w http.ResponseWriter, r *http.Request) {
	var request []string
	// Loop through headers
	for name, headers := range r.Header {
		name = strings.ToLower(name)
		for _, h := range headers {
			request = append(request, fmt.Sprintf("%v: %v", name, h))
		}
	}
	js, err := json.Marshal(request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

func (a API) routes(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		a.getRoutes(w, r)
	case "POST":
		// Create a new record.
		a.putRoute(w, r)
	case "PUT":
		// Update an existing record.
		a.putRoute(w, r)
	case "DELETE":
		// Remove the record.
		a.delRoute(w, r)
	default:
		// Give an error message.
	}
}

func (a API) putRoute(w http.ResponseWriter, r *http.Request) {
	var route types.Route
	if r.Body == nil {
		http.Error(w, "Please send a request body", 400)
		return
	}
	err := json.NewDecoder(r.Body).Decode(&route)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	a.sync.UpdateRoute(route)
}

func (a API) delRoute(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/v1/routes/")

	log.Printf("Trying to DELETE route for %s", id)

	if status, err := a.sync.DeleteRoute(id); err != nil {
		http.Error(w, err.Error(), status)
		return
	}
	w.WriteHeader(http.StatusNoContent)
	return
}

func (a API) getRoutes(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/v1/routes/")

	var js []byte
	var err error
	routes := a.sync.Routes()

	if id == "" {
		js, err = json.Marshal(a.sync.Routes())
	} else if route := routes[id]; route.ID != "" {
		log.Printf("Trying to GET route for %s", id)
		js, err = json.Marshal(a.sync.Routes()[id])
	} else {
		http.NotFound(w, r)
		return
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

// Mux returns the path multiplexer for the API
func (a API) Mux() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/echo", a.echo)
	mux.HandleFunc("/api/v1/routes/", a.routes)
	return mux
}

// New returns a new instance of the proxy
func New(sync *routes.Synchronizer) API {
	return API{
		sync: sync,
	}
}
