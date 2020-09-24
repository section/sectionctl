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
	pattern := "section-cli-api-auth-credential-" + strings.ReplaceAll(t.Name(), "/", "_")
	file, err := ioutil.TempFile("", pattern)
	if err != nil {
		t.FailNow()
	}
	return file.Name()
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
	m, u, p, err := PromptForCredential()

	// Test
	assert.NoError(err)
	assert.Equal(machine, m)
	assert.Equal(username, u)
	assert.Equal(password, p)
}

func TestAPIAuthWriteCredentialCreatesFile(t *testing.T) {
	assert := assert.New(t)

	testCases := []string{
		newCredentialTempfile(t),
		newCredentialTempfile(t) + ".new",
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

			u, p, err := GetBasicAuth()
			assert.NoError(err)
			assert.Equal(u, username)
			assert.Equal(p, password)
		})
	}
}

func TestAPIAuthGetBasicAuthReturnsCredential(t *testing.T) {
	assert := assert.New(t)

	// Setup
	CredentialPath = filepath.Join("testdata", "valid-credentials")

	// Test
	assert.True(IsCredentialRecorded())

	u, p, err := GetBasicAuth()
	assert.NoError(err)
	assert.Equal("ada@section.example", u)
	assert.Equal("v4l1ds3cr3t", p)
}

func TestAPIAuthGetBasicAuthReturnsErrorIfCredentialInvalid(t *testing.T) {
	assert := assert.New(t)

	// Setup
	CredentialPath = filepath.Join("testdata", "empty-file")

	// Invoke
	_, _, err := GetBasicAuth()

	// Test
	assert.Error(err)
	assert.False(IsCredentialRecorded())
}
