package api

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAPIClientSetsUserAgent(t *testing.T) {
	assert := assert.New(t)

	// Setup
	var userAgent string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userAgent = r.Header["User-Agent"][0]
		w.WriteHeader(http.StatusOK)
	}))

	u, err := url.Parse(ts.URL)
	assert.NoError(err)

	// Invoke
	_, err = request(http.MethodGet, u.String(), nil)
	assert.NoError(err)

	// Test
	assert.Regexp("^section-cli (.+)$", userAgent)
}