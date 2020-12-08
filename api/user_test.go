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

func TestAPIUserReturnsRecord(t *testing.T) {
	assert := assert.New(t)

	// Setup
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, string(helperLoadBytes(t, "user.with_success.json")))
	}))

	ur, err := url.Parse(ts.URL)
	assert.NoError(err)
	PrefixURI = ur

	auth.WriteCredential(ur.Host, "s3cr3t")

	// Invoke
	u, err := CurrentUser()

	// Test
	assert.NoError(err)
	assert.Equal(u.ID, 1)
	assert.Equal(u.Email, "ada@lovelace.example")
	assert.Equal(u.FirstName, "Ada")
	assert.Equal(u.LastName, "Lovelace")
	assert.Equal(u.CompanyName, "Example Corp")
	assert.Equal(u.PhoneNumber, "+13125550690")
}

func TestAPIUserHandlesErrors(t *testing.T) {
	assert := assert.New(t)

	// Setup
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTeapot)
		fmt.Fprintf(w, "short and stout")
	}))

	ur, err := url.Parse(ts.URL)
	assert.NoError(err)
	PrefixURI = ur

	auth.WriteCredential(ur.Host, "s3cr3t")

	// Invoke
	u, err := CurrentUser()

	// Test
	assert.Error(err)
	assert.Equal(u, User{})
	assert.Contains(err.Error(), "418")
	assert.Contains(err.Error(), "short and stout")
}
