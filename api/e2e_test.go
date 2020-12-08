package api

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/section/sectionctl/api/auth"
	"github.com/stretchr/testify/assert"
)

func TestAPIAuthCanWriteReadAndUseCredential(t *testing.T) {
	assert := assert.New(t)

	// Setup
	var (
		endpoint string
		token    string
	)

	// Run a mock server we'll use later
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		to := r.Header.Get("section-token")
		assert.NotEmpty(to)

		if to == token {
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
	token = "s3cr3t"

	// Write credential
	err = auth.WriteCredential(endpoint, token)
	assert.NoError(err)

	// Read Credential
	to, err := auth.GetCredential(endpoint)
	assert.NoError(err)
	assert.Equal(token, to)

	// Use credential
	usr, err := CurrentUser()
	assert.NoError(err)
	assert.NotEqual(usr, User{})
}
