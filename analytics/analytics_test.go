package analytics

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAnalyticsSubmitPostsToHeap(t *testing.T) {
	assert := assert.New(t)
	var called bool
	consentPath = newConsentTempfile(t)
	ConsentGiven = true
	err := WriteConsent(ConsentGiven)
	assert.NoError(err)

	// Setup
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true

		body, err := ioutil.ReadAll(r.Body)
		assert.NoError(err)
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
	err = Submit(e)

	// Test
	assert.NoError(err)
	assert.True(called)
}

func TestAnalyticsSubmitHandlesErrors(t *testing.T) {
	assert := assert.New(t)
	var called bool
	consentPath = newConsentTempfile(t)
	ConsentGiven = true
	err := WriteConsent(ConsentGiven)
	assert.NoError(err)

	// Setup
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusInternalServerError)
	}))
	HeapBaseURI = ts.URL

	// Invoke
	e := Event{
		Name:       "CLI invoked",
		Properties: map[string]string{"Subcommand": "apps list"},
	}
	err = Submit(e)

	// Test
	assert.Error(err)
	assert.True(called)
}
