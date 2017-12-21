package ip

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

type testHandler struct {
	responseCode int
	response     string
}

func (t *testHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(t.responseCode)
	w.Write([]byte(t.response))
}

func TestResolve(t *testing.T) {
	r := NewIPify()
	ips, err := r.Resolve()
	assert.NoError(t, err)
	assert.NotNil(t, ips)
	assert.Equal(t, 1, len(ips))
	assert.Equal(t, 4, len(strings.Split(ips[0], ".")))
}

func TestResolveErrors(t *testing.T) {
	r := IPify{URL: "file:///tmp/this/file/should/not/exist"}
	ips, err := r.Resolve()
	assert.Error(t, err)
	assert.Nil(t, ips)

	h := &testHandler{
		http.StatusCreated,
		"",
	}
	s := httptest.NewServer(h)
	defer s.Close()

	r = IPify{URL: s.URL}
	ips, err = r.Resolve()
	assert.Error(t, err)
	assert.Nil(t, ips)
	assert.Contains(t, err.Error(), "200")
	assert.Contains(t, err.Error(), "201")

	h.responseCode = http.StatusOK
	h.response = "something not json"
	ips, err = r.Resolve()
	assert.Error(t, err)
	assert.Nil(t, ips)

}
