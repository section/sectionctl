package commands

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/creack/pty"
	"github.com/section/section-cli/api"
	"github.com/section/section-cli/api/auth"
	"github.com/stretchr/testify/assert"
)

func newCredentialTempfile(t *testing.T) string {
	pattern := "section-cli-api-auth-credential-" + strings.ReplaceAll(t.Name(), "/", "_")
	file, err := ioutil.TempFile("", pattern)
	if err != nil {
		t.FailNow()
	}
	return file.Name()
}

func TestCommandsLoginValidatesGoodCredentials(t *testing.T) {
	assert := assert.New(t)

	// Setup
	username := "grace@hopper.example"
	password := "supers3cr3t"
	input := username + "\n" + password + "\n"
	c := exec.Command("echo", input)
	tty, err := pty.Start(c)
	auth.TTY = tty
	assert.NoError(err)
	defer func() { auth.TTY = os.Stdin }()
	auth.CredentialPath = newCredentialTempfile(t)

	var called bool
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "{}")
	}))

	ur, err := url.Parse(ts.URL)
	assert.NoError(err)
	api.PrefixURI = ur.String()

	// Invoke
	cmd := LoginCmd{}
	err = cmd.Run()

	// Test
	assert.NoError(err)
	assert.True(called)
}

func TestCommandsLoginValidatesBadCredentials(t *testing.T) {
	assert := assert.New(t)

	// Setup
	username := "grace@hopper.example"
	password := "b4ds3cr3t"
	input := username + "\n" + password + "\n"
	c := exec.Command("echo", input)
	tty, err := pty.Start(c)
	auth.TTY = tty
	assert.NoError(err)
	defer func() { auth.TTY = os.Stdin }()
	auth.CredentialPath = newCredentialTempfile(t)

	var called bool
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusUnauthorized)
	}))

	ur, err := url.Parse(ts.URL)
	assert.NoError(err)
	api.PrefixURI = ur.String()

	// Invoke
	cmd := LoginCmd{}
	err = cmd.Run()

	// Test
	assert.Error(err)
	assert.True(called)
	assert.Contains(err.Error(), "invalid credentials")
}
