package docker

import (
	"fmt"
	"strings"
)

// Rule holds the description of a traefik rule: https://docs.traefik.io/basics/#matchers
type Rule struct {
	Matcher string
	Values  []string
}

// ParseRule parses a string to extract structured rule description
func ParseRule(encoded string) (*Rule, error) {
	values := strings.SplitN(encoded, ":", 2)
	if len(values) != 2 {
		return nil, fmt.Errorf("invalid encoded rule %s, rule must have format <Matcher>: <values>", encoded)
	}

	rule := &Rule{}
	rule.Matcher = strings.Trim(values[0], "\t ")
	for _, pattern := range strings.Split(values[1], ",") {
		rule.Values = append(rule.Values, strings.Trim(pattern, " \t"))
	}
	return rule, nil
}

// ParseRules parses a string to extract structures rules
func ParseRules(encoded string) (map[string]*Rule, error) {
	rules := map[string]*Rule{}
	for _, e := range strings.Split(encoded, ";") {
		rule, err := ParseRule(e)
		if err != nil {
			return nil, err
		}
		rules[rule.Matcher] = rule
	}
	return rules, nil
}

// ExtractTraefikDomainsFromLabels parses all labels and extracts domains to be served
// It currently supports traefik labels and can only extract static in the Host rule
// Extrapolating the values of HostRegexp is currently not supported as it can lead
// to infinite numbers of supported domains
func ExtractTraefikDomainsFromLabels(labels map[string]string) ([]string, error) {
	encoded, ok := labels["traefik.frontend.rule"]
	if ok {
		rules, err := ParseRules(encoded)
		if err != nil {
			return nil, err
		}
		if host, ok := rules["Host"]; ok {
			return host.Values, nil
		}
	}
	return []string{}, nil

}
