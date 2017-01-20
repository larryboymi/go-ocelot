package poller

import (
	"strconv"

	"github.com/docker/docker/api/types/filters"
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
	filter := filters.NewArgs()
	filter.Add("label", "ingress=true")
	filter.Add("label", "ingressport")

	services := p.client.GetServices(filter)
	return parseRoutes(services)
}

func parseRoutes(services []swarm.Service) []types.Route {
	var serviceList []types.Route

	for _, s := range services {
		if port, err := strconv.Atoi(s.Spec.Annotations.Labels["ingressport"]); err == nil {
			serviceList = append(serviceList, types.Route{
				ID:         s.Spec.Annotations.Name,
				TargetPort: port,
			})
		}
	}
	return serviceList
}

//New poller
func New() Poller {
	client := docker.New()
	return Poller(&dockerWrapper{client: client})
}
