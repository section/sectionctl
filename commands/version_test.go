package commands

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/section/sectionctl/version"
	"github.com/stretchr/testify/assert"
)

func TestCommandsVersionSaysIfNewVersionAvailable(t *testing.T) {
	assert := assert.New(t)

	// Setup
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, string(helperLoadBytes(t, "version/latest_release.with_success.json")))
	}))

	ur, err := url.Parse(ts.URL)
	assert.NoError(err)

	version.Version = "v1.5.0"
	c := VersionCmd{
		out:              &bytes.Buffer{},
		LatestReleaseURL: ur,
		Timeout:          2 * time.Second,
	}

	// Invoke
	c.Run()

	// Test
	assert.Regexp("new version of sectionctl", c.out)
}

func TestCommandsVersionSkipsInDev(t *testing.T) {
	assert := assert.New(t)

	// Setup
	version.Version = "dev"
	c := VersionCmd{
		out: &bytes.Buffer{},
	}

	// Invoke
	err := c.Run()

	// Test
	assert.NoError(err)
	assert.Regexp("dev", c.out)
	assert.NotRegexp("new version of sectionctl", c.out)
}

func TestCommandsVersionHandlesVersionCheckTimeout(t *testing.T) {
	assert := assert.New(t)

	// Setup
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(10 * time.Second)
	}))

	ur, err := url.Parse(ts.URL)
	assert.NoError(err)

	version.Version = "v1.5.0"
	c := VersionCmd{
		out:              &bytes.Buffer{},
		LatestReleaseURL: ur,
		Timeout:          0 * time.Second,
	}

	// Invoke
	err = c.Run()

	// Test
	assert.Error(err)
	assert.ErrorIs(err, context.DeadlineExceeded)
}

func TestCommandsVersionHandlesVersionBadResponses(t *testing.T) {}
