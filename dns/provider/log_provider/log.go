package logProvider

import (
	"strings"

	"github.com/tjamet/mohotani/logger"
)

// Log is a DNS updater that prints new values on the standard logger
type Log struct {
	Logger logger.Logger
}

// Update updates DNS records for the given domain
func (l *Log) Update(domain string, ips ...string) error {
	l.Logger.Printf("Update domain IPS: %s: %s", domain, strings.Join(ips, ", "))
	return nil
}
