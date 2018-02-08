package docker

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
)

// Lister holds a docker client
type Lister struct {
	Client *client.Client
}

func logErrors(t string, c <-chan error) {
	for e := range c {
		log.Printf("warning: Got an error while getting %s events: %s", t, e.Error())
	}
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

// EventTicker proxies a ticker and adds ticks for each container and service events
func (d *Lister) EventTicker(c <-chan time.Time) <-chan time.Time {
	f := filters.NewArgs()
	f.Add("type", "service")
	serviceMessages, serviceErrors := d.Client.Events(context.Background(), types.EventsOptions{Filters: f})
	f = filters.NewArgs()
	f.Add("type", "container")
	containerMessages, containerErrors := d.Client.Events(context.Background(), types.EventsOptions{Filters: f})
	go logErrors("service", serviceErrors)
	go logErrors("container", containerErrors)
	o := make(chan time.Time)
	go func() {
		log.Println("starting docker event ticker")
		for {
			select {
			case m := <-serviceMessages:
				o <- time.Now()
				fmt.Printf("Handled event %s: %s %s\n", m.ID, m.Type, m.Status)
			case m := <-containerMessages:
				o <- time.Now()
				fmt.Printf("Handled event %s: %s %s\n", m.ID, m.Type, m.Status)
			case t := <-c:
				o <- t
			}
		}
	}()
	return o
}
