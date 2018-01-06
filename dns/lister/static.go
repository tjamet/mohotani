package lister

// Static is a resolver that have static IPs
type Static struct {
	Domains []string
}

// List implements the Lister interface
func (s *Static) List() ([]string, error) {
	return s.Domains, nil
}
