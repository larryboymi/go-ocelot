package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/larryboymi/go-ocelot/routes"
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
	case "PUT":
		// Update an existing record.
	case "DELETE":
		// Remove the record.
	default:
		// Give an error message.
	}
}
func (a API) getRoutes(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/v1/routes/")

	var js []byte
	var err error
	if id != r.URL.Path && id != "" {
		log.Printf("Trying to GET route for %s", id)
		js, err = json.Marshal(a.sync.Routes()[id])
	} else {
		js, err = json.Marshal(a.sync.Routes())
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
