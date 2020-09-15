package commands

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCommandVersionPrintsIfNewVersionAvailable(t *testing.T) {
}

func TestCommandVersionPrintsIfNoNewVersionAvailable(t *testing.T) {
	assert := assert.New(t)

	// Setup
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `{"latest_version":"%s"}`, "0.0.1")
	}))
	defer ts.Close()

	VersionCheckURL = ts.URL
	c := VersionCmd{}
	v, err := c.checkVersion()

	// Test
	assert.NoError(err)
	assert.Equal(v, "0.0.1")
}

func TestCommandVersioncheckVersionHandlesRequestFailure(t *testing.T) {
}
