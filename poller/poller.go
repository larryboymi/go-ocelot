package poller

import (
	"github.com/docker/docker/api/types/swarm"

	"github.com/larryboymi/go-ocelot/docker"
	"github.com/larryboymi/go-ocelot/types"
)

// Poller for routes
type Poller interface {
	Load() []types.Route
}

type dockerWrapper struct {
	client docker.Client
}

// LoadAll queries docker for its service and parses the ones with correct labels
func (p *dockerWrapper) Load() []types.Route {
	filters := map[string]string{"label": "ingress=true"}
	services := p.client.GetServices(filters)
	return parseRoutes(services)
}

func parseRoutes(services []swarm.Service) []types.Route {
	var serviceList []types.Route

	for _, s := range services {
		serviceList = append(serviceList, types.Route{
			ID:        s.Spec.Annotations.Name,
			TargetURL: s.Spec.Annotations.Name,
		})
	}
	return serviceList
}

//New poller
func New() Poller {
	client := docker.New()
	return Poller(&dockerWrapper{client: client})
}
