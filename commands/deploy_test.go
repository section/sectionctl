package commands

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCommandsDeployBuildFilelistIgnoresFiles(t *testing.T) {
	assert := assert.New(t)

	// Setup
	dir := filepath.Join("testdata", "deploy", "tree")
	ignores := []string{".git", "foo"}

	// Invoke
	paths, err := BuildFilelist(dir, ignores)

	// Test
	assert.NoError(err)
	assert.Greater(len(paths), 0)
	for _, p := range paths {
		for _, i := range ignores {
			assert.NotContains(p, i)
		}
	}
}

func TestCommandsDeployBuildFilelistErrorsOnNonExistentDirectory(t *testing.T) {
	assert := assert.New(t)

	// Setup
	testCases := []string{
		filepath.Join("testdata", "deploy", "non-existent-tree"),
		filepath.Join("testdata", "deploy", "file"),
	}
	var ignores []string

	for _, tc := range testCases {
		t.Run(tc, func(t *testing.T) {
			// Invoke
			paths, err := BuildFilelist(tc, ignores)

			// Test
			assert.Error(err)
			assert.Zero(len(paths))
		})
	}
}

func TestCommandsDeployUploadsTarball(t *testing.T) {
	assert := assert.New(t)

	// Setup
	var called bool
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		called = true
	}))

	dir := filepath.Join("testdata", "deploy", "tree")
	url, err := url.Parse(ts.URL)
	assert.NoError(err)

	// Invoke
	c := DeployCmd{Directory: dir, ServerURL: url}
	err = c.Run()

	assert.NoError(err)
	assert.True(called)
}
