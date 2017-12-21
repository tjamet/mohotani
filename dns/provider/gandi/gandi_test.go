package gandi

import (
	"fmt"
	"testing"

	gdomain "github.com/prasmussen/gandi-api/live_dns/domain"
	grecord "github.com/prasmussen/gandi-api/live_dns/record"
	"github.com/stretchr/testify/assert"
)

type testDomainClient struct {
	domains []*gdomain.DomainInfoBase
	err     error
	record  *testRecordClient
}
type testRecordClient struct {
	status        grecord.Status
	err           error
	updatedValues grecord.RecordInfo
	args          []string
}

func (t *testDomainClient) List() ([]*gdomain.DomainInfoBase, error) {
	return t.domains, t.err
}
func (t *testDomainClient) Records(string) grecord.Manager {
	return t.record
}
func (t *testRecordClient) Create(recordInfo grecord.RecordInfo, args ...string) (*grecord.Status, error) {
	return nil, nil
}
func (t *testRecordClient) Update(recordInfo grecord.RecordInfo, args ...string) (*grecord.Status, error) {
	t.updatedValues = recordInfo
	t.args = args
	return &t.status, t.err
}
func (t *testRecordClient) List(args ...string) ([]*grecord.RecordInfo, error) {
	return nil, nil
}
func (t *testRecordClient) Delete(args ...string) error {
	return nil
}

func TestUpdateError(t *testing.T) {
	c := testDomainClient{
		domains: nil,
		err:     fmt.Errorf("test error"),
		record:  &testRecordClient{},
	}
	gandi := Gandi{
		&c,
	}
	err := gandi.Update("test.example.com", "127.0.0.1")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "test.example.com")
	assert.Contains(t, err.Error(), "base domain")

	c.domains = []*gdomain.DomainInfoBase{
		&gdomain.DomainInfoBase{
			Fqdn: "example.cat",
		},
		&gdomain.DomainInfoBase{
			Fqdn: "example.frcat",
		},
	}
	c.err = nil
	err = gandi.Update("test.example.com", "127.0.0.1")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "example.cat")
	assert.Contains(t, err.Error(), "example.fr")
	assert.Contains(t, err.Error(), "test.example.com")
	assert.Contains(t, err.Error(), "base domain")

	c.domains = []*gdomain.DomainInfoBase{
		&gdomain.DomainInfoBase{
			Fqdn: "example.cat",
		},
		&gdomain.DomainInfoBase{
			Fqdn: "example.com",
		},
	}
	err = gandi.Update("example.com", "127.0.0.1")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "example.com")
	assert.Contains(t, err.Error(), "root domain")

	c.record.err = fmt.Errorf("test error")
	err = gandi.Update("test.example.com", "127.0.0.1")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "test.example.com")
	assert.Contains(t, err.Error(), "127.0.0.1")
	assert.Contains(t, err.Error(), "test error")

	c.record.updatedValues = grecord.RecordInfo{}
	c.record.err = nil
	err = gandi.Update("test.example.com", "127.0.0.1")
	assert.NoError(t, err)
	assert.Equal(t, grecord.RecordInfo{Values: []string{"127.0.0.1"}}, c.record.updatedValues)
	assert.Equal(t, []string{"test", "A"}, c.record.args)
}

func TestNew(t *testing.T) {
	g := New("api key").domainAccessor.(*gdomain.Domain)
	assert.Equal(t, "api key", g.Key)
}
