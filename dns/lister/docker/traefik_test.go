package docker

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseRule(t *testing.T) {
	tests := []struct {
		serialized string
		matcher    string
		values     []string
	}{
		{serialized: "Headers: Content-Type, application/json", matcher: "Headers", values: []string{"Content-Type", "application/json"}},
		{serialized: "Headers: Content-Type, application/(text/json)	", matcher: "Headers", values: []string{"Content-Type", "application/(text/json)"}},
		{serialized: "Host: traefik.io, www.traefik.io	", matcher: "Host", values: []string{"traefik.io", "www.traefik.io"}},
		{serialized: "HostRegexp: traefik.io, {subdomain:[a-z]+}.traefik.io	", matcher: "HostRegexp", values: []string{"traefik.io", "{subdomain:[a-z]+}.traefik.io"}},
		{serialized: "Method: GET, POST, PUT", matcher: "Method", values: []string{"GET", "POST", "PUT"}},
		{serialized: "Path: /products/, /articles/{category}/{id:[0-9]+}	", matcher: "Path", values: []string{"/products/", "/articles/{category}/{id:[0-9]+}"}},
		{serialized: "PathStrip: /products/", matcher: "PathStrip", values: []string{"/products/"}},
		{serialized: "PathStripRegex: /articles/{category}/{id:[0-9]+}	", matcher: "PathStripRegex", values: []string{"/articles/{category}/{id:[0-9]+}"}},
		{serialized: "PathPrefix: /products/, /articles/{category}/{id:[0-9]+}	", matcher: "PathPrefix", values: []string{"/products/", "/articles/{category}/{id:[0-9]+}"}},
		{serialized: "PathPrefixStrip: /products/", matcher: "PathPrefixStrip", values: []string{"/products/"}},
		{serialized: "PathPrefixStripRegex: /articles/{category}/{id:[0-9]+}	", matcher: "PathPrefixStripRegex", values: []string{"/articles/{category}/{id:[0-9]+}"}},
		{serialized: "Query: foo=bar, bar=baz", matcher: "Query", values: []string{"foo=bar", "bar=baz"}},
	}
	for _, test := range tests {
		t.Run("With serialized:"+test.serialized, func(t *testing.T) {
			r, err := ParseRule(test.serialized)
			assert.NoError(t, err)
			assert.NotNil(t, r)
			if r != nil {
				assert.Equal(t, test.matcher, r.Matcher)
				assert.Equal(t, test.values, r.Values)
			}
		})
	}
}

func TestParseRules(t *testing.T) {
	rules, err := ParseRules("Host: traefik.io, www.traefik.io;	PathPrefix: /products/, /articles/{category}/{id:[0-9]+}")
	assert.NoError(t, err)
	assert.Contains(t, rules, "Host")
	assert.Contains(t, rules, "PathPrefix")
	assert.NotNil(t, rules["Host"])
	assert.NotNil(t, rules["PathPrefix"])
}

func TestExtractDomainsFromLabels(t *testing.T) {
	domains, err := ExtractTraefikDomainsFromLabels(map[string]string{"traefik.frontend.rule": "Host: traefik.io, www.traefik.io;	PathPrefix: /products/, /articles/{category}/{id:[0-9]+}"})
	assert.NoError(t, err)
	assert.Equal(t, []string{"traefik.io", "www.traefik.io"}, domains)

	domains, err = ExtractTraefikDomainsFromLabels(map[string]string{"traefik": "Host: traefik.io, www.traefik.io;	PathPrefix: /products/, /articles/{category}/{id:[0-9]+}"})
	assert.NoError(t, err)
	assert.Equal(t, []string{}, domains)

	domains, err = ExtractTraefikDomainsFromLabels(map[string]string{"traefik.frontend.rule": "Host"})
	assert.Error(t, err)
	assert.Nil(t, domains)
}
