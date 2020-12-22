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
	assert.Contains(err.Error(), "invalid credentials")
}
