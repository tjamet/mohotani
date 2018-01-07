package logProvider

import (
	"log"
	"strings"
)

// Log is a DNS updater that prints new values on the standard logger
type Log struct{}

// Update updates DNS records for the given domain
func (l *Log) Update(domain string, ips ...string) error {
	log.Printf("Update domain IPS: %s: %s", domain, strings.Join(ips, ", "))
	return nil
}
