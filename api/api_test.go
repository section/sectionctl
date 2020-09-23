package api

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"testing"

	"github.com/section/section-cli/api/auth"
	"github.com/stretchr/testify/assert"
)

func helperLoadBytes(t *testing.T, name string) []byte {
	path := filepath.Join("testdata", name) // relative path
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		assert.FailNow(t, err.Error())
	}
	return bytes
}

func TestAPIClientSetsUserAgent(t *testing.T) {
	assert := assert.New(t)

	// Setup
	var userAgent string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userAgent = r.Header["User-Agent"][0]
		w.WriteHeader(http.StatusOK)
	}))

	u, err := url.Parse(ts.URL)
	assert.NoError(err)

	auth.CredentialPath = filepath.Join("auth", "testdata", "valid-credentials")

	// Invoke
	_, err = request(http.MethodGet, u.String(), nil)
	assert.NoError(err)

	// Test
	assert.Regexp("^section-cli (.+)$", userAgent)
}
