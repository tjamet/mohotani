package lister

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRListStatic(t *testing.T) {
	s := Static{[]string{"www.thib-o.eu"}}
	domains, err := s.List()
	assert.NoError(t, err)
	assert.Equal(t, []string{"www.thib-o.eu"}, domains)

	for _, expectedDomains := range [][]string{
		{"www2.thib-o.eu", "www3.thib-o.eu"},
		{"test.example.org", "test.example.com"},
	} {
		s.Domains = expectedDomains
		domains, err = s.List()
		assert.NoError(t, err)
		assert.Equal(t, expectedDomains, domains)
	}
}
