package docker

import (
	"context"
	"fmt"
	"log"

	"github.com/docker/docker/api/types"
	"github.com/moby/moby/client"
)

// Lister holds a docker client
type Lister struct {
	Client *client.Client
}

// List implements the Lister interface to list required domains
// for all containers and services
func (d *Lister) List() ([]string, error) {
	domains := map[string]interface{}{}
	containers, err := d.Client.ContainerList(context.Background(), types.ContainerListOptions{})
	if err != nil {
		return nil, err
	}
	for _, container := range containers {
		newDomains, err := ExtractTraefikDomainsFromLabels(container.Labels)
		if err != nil {
			log.Printf("Failed to extract domain names for container %s, %s", container.Names, err.Error())
		}
		fmt.Println(container.Labels, container.ID)
		for _, domain := range newDomains {
			domains[domain] = nil
		}
	}
	_, err = d.Client.SwarmInspect(context.Background())
	if err == nil {
		services, err := d.Client.ServiceList(context.Background(), types.ServiceListOptions{})
		if err != nil {
			return nil, err
		}
		for _, service := range services {
			newDomains, err := ExtractTraefikDomainsFromLabels(service.Spec.Labels)
			if err != nil {
				log.Printf("Failed to extract domain names for container %s, %s", service.Spec.Name, err.Error())
			}
			for _, domain := range newDomains {
				domains[domain] = nil
			}
		}
	}
	domainList := []string{}
	for key := range domains {
		domainList = append(domainList, key)
	}
	return domainList, nil
}
