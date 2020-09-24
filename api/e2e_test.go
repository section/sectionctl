package api

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/section/section-cli/api/auth"
	"github.com/stretchr/testify/assert"
)

func TestAPIAuthCanWriteReadAndUseCredential(t *testing.T) {
	assert := assert.New(t)

	// Setup
	var (
		username string
		password string
		endpoint string
	)

	// Run a mock server we'll use later
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, pass, ok := r.BasicAuth()

		assert.True(ok)

		if user == username && pass == password {
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, string(helperLoadBytes(t, "user.with_success.json")))
		} else {
			w.WriteHeader(http.StatusForbidden)
		}
	}))

	ur, err := url.Parse(ts.URL)
	assert.NoError(err)
	PrefixURI = ur
	endpoint = ur.Host
	username = "grace@hopper.example"
	password = "s3cr3t"

	auth.CredentialPath = newCredentialTempfile(t)

	// Write credential
	err = auth.WriteCredential(endpoint, username, password)
	assert.NoError(err)

	// Read Credential
	u, p, err := auth.GetCredential(endpoint)
	assert.NoError(err)
	assert.Equal(username, u)
	assert.Equal(password, p)

	// Use credential
	usr, err := CurrentUser()
	assert.NoError(err)
	assert.NotEqual(usr, User{})
}
