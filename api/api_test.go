package api

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"strings"
	"testing"

	"github.com/section/sectionctl/api/auth"
	"github.com/stretchr/testify/assert"
)

func newCredentialTempfile(t *testing.T) string {
	pattern := "sectionctl-api-auth-credential-" + strings.ReplaceAll(t.Name(), "/", "_")
	file, err := ioutil.TempFile("", pattern)
	if err != nil {
		t.FailNow()
	}
	return file.Name()
}

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

	auth.CredentialPath = newCredentialTempfile(t)
	auth.WriteCredential(u.Host, "foo", "bar")

	// Invoke
	_, err = request(http.MethodGet, *u, nil)
	assert.NoError(err)

	// Test
	assert.Regexp("^sectionctl (.+)$", userAgent)
}

func TestPrettyTxIDErrorPrintsApertureTxID(t *testing.T) {
	assert := assert.New(t)

	// Invoke
	resp := http.Response{Status: "500 Internal Server Error", Header: map[string][]string{"Aperture-Tx-Id": []string{"12345"}}}
	err := prettyTxIDError(&resp)

	// Test
	assert.Error(err)
	assert.Regexp("transaction ID", err)
	assert.Regexp(resp.Header["Aperture-Tx-Id"][0], err)
}

func TestPrettyTxIDErrorHandlesNoApertureTxIDHeader(t *testing.T) {
	assert := assert.New(t)

	// Invoke
	resp := http.Response{Status: "500 Internal Server Error"}
	assert.NotPanics(func() { prettyTxIDError(&resp) })
	err := prettyTxIDError(&resp)

	// Test
	assert.Error(err)
	assert.NotRegexp("transaction ID", err)
}

func TestAPIClientUsesCredentialsIfSpecified(t *testing.T) {
	assert := assert.New(t)

	// Setup
	var (
		username string
		password string
	)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		u, p, ok := r.BasicAuth()
		username = u
		password = p
		assert.True(ok)
		w.WriteHeader(http.StatusOK)
	}))

	Username = "alice"
	Token = "s3cr3t"

	u, err := url.Parse(ts.URL)
	assert.NoError(err)

	// Invoke
	resp, err := request(http.MethodGet, *u, nil)
	assert.NoError(err)

	// Test
	assert.Equal(resp.StatusCode, http.StatusOK)
	assert.Equal(Username, username)
	assert.Equal(Token, password)

	// Teardown
	Username = ""
	Token = ""
}
