package gandi

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
	gclient "github.com/prasmussen/gandi-api/client"
	gdomain "github.com/prasmussen/gandi-api/live_dns/domain"
	grecord "github.com/prasmussen/gandi-api/live_dns/record"
)

type domainLister interface {
	List() ([]*gdomain.InfoBase, error)
}

type recordAccessor interface {
	Records(string) grecord.Manager
}

type domainAccessor interface {
	domainLister
	recordAccessor
}

// Gandi implements the updater interface for gandi liveDNS API
type Gandi struct {
	domainAccessor
}

// New instanciates a new gandi updater for gandi liveDNS A records
func New(key string) *Gandi {
	c := gclient.New(key, gclient.LiveDNS)
	return &Gandi{
		gdomain.New(c),
	}
}

// Update updates DNS records for the given domain
func (g *Gandi) Update(domain string, ips ...string) error {
	domains, err := g.domainAccessor.List()
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("unable to find base domain for '%s'", domain))
	}
	var baseDomain *gdomain.InfoBase
	for _, d := range domains {
		if strings.HasSuffix(domain, d.Fqdn) {
			baseDomain = d
			break
		}
	}
	if baseDomain == nil {
		availableDomains := []string{}
		for _, d := range domains {
			availableDomains = append(availableDomains, d.Fqdn)
		}
		return fmt.Errorf("no base domain found for '%s' using gandi API, found domains: [%s]", domain, strings.Join(availableDomains, ","))
	}
	r := strings.TrimSuffix(strings.TrimSuffix(domain, baseDomain.Fqdn), ".")
	if len(r) == 0 {
		return fmt.Errorf("unable to update the root domain of %s", domain)
	}
	_, err = g.domainAccessor.Records(baseDomain.Fqdn).Update(grecord.Info{Values: ips}, r, grecord.A)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("unable to update record infos for domain '%s' with ips %s", domain, strings.Join(ips, ",")))
	}
	return err
}
