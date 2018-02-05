package updater

import (
	"strings"

	"github.com/tjamet/mohotani/dns/provider"
	"github.com/tjamet/mohotani/listener"
)

// Updater his the structure holding the setup for automatic dns records update
type Updater struct {
	Updater        provider.Updater
	IPListener     listener.Listener
	DomainListener listener.Listener
	Logger         listener.Logger
}

func (u *Updater) apply(domains, IPs []string) {
	if domains != nil && IPs != nil {
		for _, domain := range domains {
			err := u.Updater.Update(domain, IPs...)
			if err != nil {
				u.Logger.Printf("failed to update domain %s: %s", domain, err)
			} else {
				u.Logger.Printf("updated domain %s with IPs %s", domain, strings.Join(IPs, ","))
			}
		}
	}
}

// Start applies the record registry updates in case of any change in either the IP or the domains
func (u *Updater) Start() {
	ipsChannel := make(chan []string)
	domainsChannel := make(chan []string)
	var IPs []string
	var domains []string
	go u.IPListener.Listen(ipsChannel)
	go u.DomainListener.Listen(domainsChannel)
	for {
		select {
		case IPs = <-ipsChannel:
			u.apply(domains, IPs)
		case domains = <-domainsChannel:
			u.apply(domains, IPs)
		}
	}
}
