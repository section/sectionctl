package api

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/section/sectionctl/api/auth"
	"github.com/stretchr/testify/assert"
)

func TestApplicationEnvironmentModuleUpdateErrorsIfRequestFails(t *testing.T) {
	assert := assert.New(t)

	// Setup
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _, ok := r.BasicAuth()
		assert.True(ok)
		w.Header().Add("Aperture-Tx-Id", "400400400400.400400")
		w.WriteHeader(http.StatusBadRequest)
	}))
	url, err := url.Parse(ts.URL)
	assert.NoError(err)
	PrefixURI = url

	auth.CredentialPath = newCredentialTempfile(t)
	endpoint := url.Host
	username := "hello"
	password := "s3cr3t"
	auth.WriteCredential(endpoint, username, password)

	// Invoke
	up := EnvironmentUpdateCommand{Op: "replace", Value: map[string]string{"hello": "world"}}
	err = ApplicationEnvironmentModuleUpdate(1, 1, "production", "hello/world.json", up)

	// Test
	assert.Error(err)
}
