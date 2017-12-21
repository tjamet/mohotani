package provider

// Updater is the interface to update the A DNS records
type Updater interface {
	// Update provides the ability to update a domain name with several IPs
	Update(domain string, ips ...string) error
}
