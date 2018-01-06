package lister

// Lister defines methods an object must implement to list all required domains
type Lister interface {
	// List returns all domain names all required domains
	List() ([]string, error)
}
