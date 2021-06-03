package commands

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/section/sectionctl/api"
	"github.com/section/sectionctl/credentials"
	"github.com/stretchr/testify/assert"
	"github.com/zalando/go-keyring"
)

func TestCommandsLoginValidatesGoodCredentials(t *testing.T) {
	assert := assert.New(t)
	keyring.MockInit()

	// Setup
	var called bool
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "{}")
	}))

	ur, err := url.Parse(ts.URL)
	assert.NoError(err)
	api.PrefixURI = ur

	in := strings.NewReader("s3cr3t\n")
	var out bytes.Buffer
	cmd := LoginCmd{
		in:  in,
		out: &out,
	}

	// Invoke
	err = cmd.Run()

	// Test
	assert.NoError(err)
	assert.True(called)
}

func TestCommandsLoginValidatesBadCredentials(t *testing.T) {
	assert := assert.New(t)
	keyring.MockInit()

	// Setup
	var called bool
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusUnauthorized)
	}))

	ur, err := url.Parse(ts.URL)
	assert.NoError(err)
	api.PrefixURI = ur

	in := strings.NewReader("s3cr3t\n")
	var out bytes.Buffer
	cmd := LoginCmd{
		in:  in,
		out: &out,
	}

	// Invoke
	err = cmd.Run()

	// Test
	assert.Error(err)
	assert.True(called)
	assert.Contains(err.Error(), "unauthorized")
}

func TestCommandsLoginUsesAlreadySetAPIToken(t *testing.T) {
	assert := assert.New(t)
	keyring.MockInit()
	api.Token = "s3cr3t-" + t.Name()

	// Setup
	var called bool
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		token := r.Header.Get("section-token")
		if assert.Equal(token, api.Token) {
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, "{}")
		} else {
			w.WriteHeader(http.StatusBadRequest)
		}
	}))

	ur, err := url.Parse(ts.URL)
	assert.NoError(err)
	api.PrefixURI = ur

	cmd := LoginCmd{
		in:  &bytes.Buffer{},
		out: &bytes.Buffer{},
	}


	// Invoke
	err = cmd.Run()

	// Test
	assert.NoError(err)
	assert.True(called)

	c, err := credentials.Read(api.PrefixURI.Host)
	assert.NoError(err)
	assert.Equal(c, api.Token)
}
