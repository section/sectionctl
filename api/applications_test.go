package api

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/section/sectionctl/credentials"
	"github.com/stretchr/testify/assert"
)

func TestApplicationEnvironmentModuleUpdateSendsUpdateInArray(t *testing.T) {
	assert := assert.New(t)

	// Setup
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		to := r.Header.Get("section-token")
		assert.NotEmpty(to)
		w.Header().Add("Aperture-Tx-Id", "400400400400.400400")

		b, err := ioutil.ReadAll(r.Body)
		assert.NoError(err)
		var ups []EnvironmentUpdateCommand
		err = json.Unmarshal(b, &ups)
		assert.NoError(err)
		if assert.Equal(len(ups), 1) {
			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(http.StatusBadRequest)
		}
	}))
	url, err := url.Parse(ts.URL)
	assert.NoError(err)
	PrefixURI = url
	Token = "s3cr3t"

	// Invoke
	var ups = []EnvironmentUpdateCommand{
		EnvironmentUpdateCommand{Op: "replace", Value: map[string]string{"hello": "world"}},
	}
	err = ApplicationEnvironmentModuleUpdate(1, 1, "production", "hello/world.json", ups)

	// Test
	assert.NoError(err)
}

func TestApplicationEnvironmentModuleUpdateErrorsIfRequestFails(t *testing.T) {
	assert := assert.New(t)

	// Setup
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		to := r.Header.Get("section-token")
		assert.NotEmpty(to)
		w.Header().Add("Aperture-Tx-Id", "400400400400.400400")
		w.WriteHeader(http.StatusBadRequest)
	}))
	url, err := url.Parse(ts.URL)
	assert.NoError(err)
	PrefixURI = url

	endpoint := url.Host
	token := "s3cr3t"
	credentials.Write(endpoint, token)

	// Invoke
	var ups = []EnvironmentUpdateCommand{
		EnvironmentUpdateCommand{Op: "replace", Value: map[string]string{"hello": "world"}},
	}
	err = ApplicationEnvironmentModuleUpdate(1, 1, "production", "hello/world.json", ups)

	// Test
	assert.Error(err)
}
