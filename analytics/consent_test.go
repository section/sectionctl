package analytics

import (
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestConsent(t *testing.T) {
	assert := assert.New(t)
	ReadConsent()
	assert.True(true)
}

func TestConsentSubmitNoopsIfNoConsent(t *testing.T) {
	assert := assert.New(t)
	var called bool
	ConsentGiven = false
	// Setup
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true

		body, _ := ioutil.ReadAll(r.Body)
		t.Logf("%s", body)
	}))
	HeapBaseURI = ts.URL

	// Invoke
	e := Event{
		Name: "CLI invoked",
		Properties: map[string]string{
			"Args":       "apps list",
			"Subcommand": "apps list",
			"Errors":     "",
		},
	}
	err := Submit(e)

	// Test
	assert.NoError(err)
	assert.False(called)
}
