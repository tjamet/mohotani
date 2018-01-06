package ip

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestResolveStatic(t *testing.T) {
	s := Static{[]string{"127.0.0.1"}}
	ips, err := s.Resolve()
	assert.NoError(t, err)
	assert.Equal(t, []string{"127.0.0.1"}, ips)

	for _, expectedIPs := range [][]string{
		{"129.168.1.1", "169.0.0.1"},
		{"129.168.1.1", "::1"},
		{"2001:0db8:0000:85a3:0000:0000:ac1f:8001", "2000:0:0:0:0:0:0:0", "129.168.1.1", "169.0.0.1"},
	} {
		s.IPs = expectedIPs
		ips, err = s.Resolve()
		assert.NoError(t, err)
		assert.Equal(t, expectedIPs, ips)
	}
}
