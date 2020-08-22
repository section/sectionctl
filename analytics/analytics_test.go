package analytics

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAnalyticsSubmitPostsToHeap(t *testing.T) {
	assert := assert.New(t)
	called := false
	// Setup
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	}))
	HeapBaseURI = ts.URL

	// Invoke
	e := Event{
		Name:       "CLI invoked",
		Properties: map[string]string{"Subcommand": "apps list"},
	}
	err := Submit(e)

	// Test
	assert.NoError(err)
	assert.True(called)
}

func TestAnalyticsSubmitHandlesErrors(t *testing.T) {
	assert := assert.New(t)
	called := false
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
	err := Submit(e)

	// Test
	assert.Error(err)
	assert.True(called)
}
