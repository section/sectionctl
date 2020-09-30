package commands

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"testing"

	"github.com/gabriel-vasile/mimetype"
	"github.com/section/sectionctl/api/auth"
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
	var req struct {
		called   bool
		username string
		password string
		body     []byte
	}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		u, p, ok := r.BasicAuth()
		assert.True(ok)
		req.called = true
		req.username = u
		req.password = p
		b, err := ioutil.ReadAll(r.Body)
		assert.NoError(err)
		req.body = b
		w.WriteHeader(http.StatusOK)
	}))

	url, err := url.Parse(ts.URL)
	assert.NoError(err)

	auth.CredentialPath = newCredentialTempfile(t)
	endpoint := url.Host
	username := "hello"
	password := "s3cr3t"
	auth.WriteCredential(endpoint, username, password)

	dir := filepath.Join("testdata", "deploy", "tree")

	// Invoke
	c := DeployCmd{Directory: dir, ServerURL: url}
	err = c.Run()

	assert.NoError(err)
	assert.True(req.called)
	assert.Equal(username, req.username)
	assert.Equal(password, req.password)
	assert.NotZero(len(req.body))

	mime, err := mimetype.DetectReader(bytes.NewReader(req.body))
	assert.NoError(err)
	assert.Equal("application/gzip", mime.String())
}
