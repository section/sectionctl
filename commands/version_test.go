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

func TestCommandsVersionHandlesBadResponses(t *testing.T) {
	assert := assert.New(t)

	// Setup
	testCases := []struct {
		ResponseBodyPath string
		ResponseStatus   int
		LatestReleaseURL *url.URL
		ExpectedError    string
	}{
		{"version/latest_release.with_bad_version_string.json", http.StatusOK, nil, "Malformed version"},
		{"version/latest_release.with_bad_json.json", http.StatusOK, nil, "unable to decode response"},
		{"version/latest_release.with_success.json", http.StatusBadRequest, nil, "bad response from GitHub"},
		{"version/latest_release.with_success.json", http.StatusOK, &url.URL{}, "unable to perform request"},
	}

	for _, tc := range testCases {
		name := fmt.Sprintf("%s-%d-%s", tc.ResponseBodyPath, tc.ResponseStatus, tc.LatestReleaseURL)
		t.Run(name, func(t *testing.T) {
			var u *url.URL

			if tc.LatestReleaseURL == nil {
				// Setup
				ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(tc.ResponseStatus)
					fmt.Fprint(w, string(helperLoadBytes(t, tc.ResponseBodyPath)))
				}))

				ur, err := url.Parse(ts.URL)
				assert.NoError(err)
				u = ur
			} else {
				u = tc.LatestReleaseURL
			}

			version.Version = "v1.5.0"
			c := VersionCmd{
				out:              &bytes.Buffer{},
				LatestReleaseURL: u,
				Timeout:          10 * time.Second,
			}
			// Invoke
			err := c.Run()

			// Test
			assert.Error(err)
			assert.Regexp(tc.ExpectedError, err)
		})
	}
}

func TestVersionCmd_Run(t *testing.T) {
	tests := []struct {
		name    string
		c       *VersionCmd
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.c.Run(); (err != nil) != tt.wantErr {
				t.Errorf("VersionCmd.Run() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
