package ip

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/pkg/errors"
)

// IPifyURL is the default address of the ipify service
const IPifyURL = "https://api.ipify.org?format=json"

// IPify implements a resolver that uses ipify.org to resolve the public IP
type IPify struct {
	URL string
}

// IP is the structure in which the ip address is de-serialized
type IP struct {
	Address string `json:"ip"`
}

// NewIPify instanciates a new IP address resolver with default address
func NewIPify() *IPify {
	return &IPify{
		URL: IPifyURL,
	}
}

// Resolve calls ipify api to get the apparent public address
func (i *IPify) Resolve() ([]string, error) {
	response, err := http.Get(i.URL)
	if err != nil {
		return nil, errors.Wrap(err, "unable to resolve current public address")
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unable to resolve current public address, unexpected http code %d from %s. expecting %d", response.StatusCode, i.URL, http.StatusOK)
	}
	ip := IP{}
	decoder := json.NewDecoder(response.Body)
	err = decoder.Decode(&ip)
	if err != nil {
		return nil, errors.Wrap(err, "unable to resolve current public address, failed to parse json")
	}
	return []string{ip.Address}, nil
}
