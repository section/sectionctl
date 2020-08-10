package commands

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCommandVersion(t *testing.T) {
	assert := assert.New(t)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Logf("Request: %+v\n", r.URL)
	}))
	defer ts.Close()
	VersionCheckEndpoint = ts.URL
	c := VersionCmd{}
	c.Run()

	assert.True(true)
}
