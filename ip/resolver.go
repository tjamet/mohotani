package ip

// Resolver defines methods an object must implement to be an IP resolver
type Resolver interface {
	// Resolve resolves the IPs the DNS should resolve
	Resolve() ([]string, error)
}
