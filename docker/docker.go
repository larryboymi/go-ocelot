package docker

import (
	"log"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/swarm"
	"github.com/docker/docker/client"
	"golang.org/x/net/context"
)

// Client takes labels and returns matching docker services
type Client interface {
	GetServices(filters.Args) []swarm.Service
}

type clientWrapper struct {
	cli *client.Client
}

// GetServices returns all Docker services matching the filter
func (c *clientWrapper) GetServices(filter filters.Args) []swarm.Service {
	defer func() {
		if r := recover(); r != nil {
			log.Print("Error getting services: ", r)
		}
	}()

	services, err := c.cli.ServiceList(context.Background(), types.ServiceListOptions{Filters: filter})
	if err != nil {
		log.Print("Error getting services: ", err)
		return []swarm.Service{}
	}

	return services
}

// New returns a new instance of the HTTP client
func New() Client {
	defaultHeaders := map[string]string{"User-Agent": "engine-api-cli-1.0"}
	cli, err := client.NewClient("unix:///var/run/docker.sock", "v1.24", nil, defaultHeaders)

	if err != nil {
		log.Panic("Failed to create Docker client: ", err)
	}

	return Client(&clientWrapper{cli: cli})
}
