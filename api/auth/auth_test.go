package auth

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/creack/pty"
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

func newCredentialTempdir(t *testing.T) string {
	pattern := "sectionctl-api-auth-credential-" + strings.ReplaceAll(t.Name(), "/", "_")
	dir, err := ioutil.TempDir("", pattern)
	if err != nil {
		t.FailNow()
	}
	return dir
}

func TestAPIAuthDetectsIfCredentialNotRecorded(t *testing.T) {
	assert := assert.New(t)

	testCases := []string{
		filepath.Join("testdata", "a-file-that-does-not-exist"),
		filepath.Join("testdata", "missing-machine"),
		filepath.Join("testdata", "missing-login"),
		filepath.Join("testdata", "missing-password"),
		filepath.Join("testdata", "missing-login-and-password"),
		filepath.Join("testdata", "zero-length-login"),
		filepath.Join("testdata", "zero-length-password"),
	}
	for _, tc := range testCases {
		t.Run(tc, func(t *testing.T) {
			// Setup
			CredentialPath = tc

			// Test
			assert.False(IsCredentialRecorded())
		})
	}
}

/*
func TestAPIAuthPromptsForCredential(t *testing.T) {
	assert := assert.New(t)

	// Setup
	machine := "aperture.section.io"
	username := "ada@section.example"
	password := "s3cr3t"
	input := username + "\n" + password + "\n"

	c := exec.Command("echo", input)
	tty, err := pty.Start(c)
	TTY = tty
	assert.NoError(err)
	defer func() { TTY = os.Stdin }()

	// Invoke
	u, p, err := PromptForCredential(machine)

	// Test
	assert.NoError(err)
	assert.Equal(username, u)
	assert.Equal(password, p)
}
*/

func TestAPIAuthWriteCredentialCreatesFile(t *testing.T) {
	assert := assert.New(t)

	testCases := []string{
		newCredentialTempfile(t),
		newCredentialTempfile(t) + ".new",
		newCredentialTempdir(t) + "/subdir/netrc",
	}
	for _, tc := range testCases {
		t.Run(tc, func(t *testing.T) {
			// Setup
			CredentialPath = tc
			machine := "aperture.section.io"
			username := "ada@lovelace.example"
			password := "s3cr3t"

			// Invoke
			err := WriteCredential(machine, username, password)

			// Test
			assert.NoError(err)
			info, err := os.Stat(CredentialPath)
			assert.NoError(err)
			assert.Equal(info.Mode().Perm(), os.FileMode(0x180)) // 0600

			u, p, err := GetCredential(machine)
			assert.NoError(err)
			assert.Equal(u, username)
			assert.Equal(p, password)
		})
	}
}

func TestAPIAuthGetCredentialReturnsBasicAuthCredential(t *testing.T) {
	assert := assert.New(t)

	// Setup
	CredentialPath = filepath.Join("testdata", "valid-credentials")

	// Invoke
	u, p, err := GetCredential("valid.example")

	// Test
	assert.NoError(err)
	assert.Equal("ada@section.example", u)
	assert.Equal("v4l1ds3cr3t", p)
}

func TestAPIAuthGetCredentialReturnsErrorIfCredentialInvalid(t *testing.T) {
	assert := assert.New(t)

	// Setup
	CredentialPath = filepath.Join("testdata", "empty-file")

	// Invoke
	_, _, err := GetCredential("foobar")

	// Test
	assert.Error(err)
	assert.False(IsCredentialRecorded())
}

func TestAPIAuthCanReadWrittenCredentials(t *testing.T) {
	assert := assert.New(t)

	// Setup
	endpoint := "127.0.0.1:8080"
	username := "grace@hopper.example"
	password := "s3cr3t"

	CredentialPath = newCredentialTempfile(t)

	// Invoke
	err := WriteCredential(endpoint, username, password)
	assert.NoError(err)

	// Test
	u, p, err := GetCredential(endpoint)
	assert.NoError(err)
	assert.Equal(username, u)
	assert.Equal(password, p)
}
