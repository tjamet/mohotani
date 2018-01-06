package ip

// Static is a resolver that have static IPs
type Static struct {
	IPs []string
}

// Resolve implements the Resolver interface
func (s *Static) Resolve() ([]string, error) {
	return s.IPs, nil
}
