package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/ocelotconsulting/go-ocelot/routes"
	"github.com/ocelotconsulting/go-ocelot/types"
)

// API is the type that handles routing api requests for go-ocelot
type API interface {
	Mux() *http.ServeMux
}

type repoWrapper struct {
	repo routes.Repository
}

func echo(w http.ResponseWriter, r *http.Request) {
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

func (repo *repoWrapper) routes(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		repo.getRoutes(w, r)
	case "POST":
		repo.putRoute(w, r)
	case "PUT":
		repo.putRoute(w, r)
	case "DELETE":
		repo.delRoute(w, r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
}

func (repo *repoWrapper) putRoute(w http.ResponseWriter, r *http.Request) {
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
	repo.repo.UpdateRoute(route)
}

func (repo *repoWrapper) delRoute(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/v1/routes/")

	log.Printf("Trying to DELETE route for %s", id)

	if status, err := repo.repo.DeleteRoute(id); err != nil {
		http.Error(w, err.Error(), status)
		return
	}
	w.WriteHeader(http.StatusNoContent)
	return
}

func (repo *repoWrapper) getRoutes(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/v1/routes/")

	var js []byte
	var err error
	routes := repo.repo.Routes()

	if id == "" {
		js, err = json.Marshal(repo.repo.Routes())
	} else if route := routes[id]; route.ID != "" {
		log.Printf("Trying to GET route for %s", id)
		js, err = json.Marshal(repo.repo.Routes()[id])
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
func (repo *repoWrapper) Mux() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/echo", echo)
	mux.HandleFunc("/api/v1/routes/", repo.routes)
	return mux
}

// New returns a new instance of the proxy
func New(repo routes.Repository) API {
	return API(&repoWrapper{
		repo: repo,
	})
}
