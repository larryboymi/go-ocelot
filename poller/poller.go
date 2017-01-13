package poller

import (
	"github.com/docker/docker/api/types/swarm"

	"github.com/larryboymi/go-ocelot/docker"
)

// Poller for routes
type Poller interface {
	Load() []Route
}

type Route struct {
	ID        string
	TargetURL string
}

type dockerWrapper struct {
	client docker.Client
}

// LoadAll queries docker for its service and parses the ones with correct labels
func (p *dockerWrapper) Load() []Route {
	filters := map[string]string{"label": "ingress=true"}
	services := p.client.GetServices(filters)
	return parseRoutes(services)
}

func parseRoutes(services []swarm.Service) []Route {
	var serviceList []Route

	for _, s := range services {
		serviceList = append(serviceList, Route{
			s.Spec.Annotations.Name,
			s.Spec.Annotations.Name,
		})
	}
	return serviceList
}

//New poller
func New() Poller {
	client := docker.New()
	return Poller(&dockerWrapper{client: client})
}
