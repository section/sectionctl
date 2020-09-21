package auth

import (
	"bytes"
	"github.com/creack/pty"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

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
			credentialPath = tc

			// Test
			assert.False(IsCredentialRecorded())
		})
	}
}

func TestAPIAuthPromptsForCredentialIfCredentialNotRecorded(t *testing.T) {
	assert := assert.New(t)

	// Setup
	credentialPath = newCredentialTempfile(t)
	t.Logf(credentialPath)

	var outbuf bytes.Buffer
	out = &outbuf

	c := exec.Command("echo", "jane@section.example\n")
	tty, err := pty.Start(c)
	assert.NoError(err)
	in = tty
	defer func() { in = os.Stdin }()

	// Invoke
	err = PromptForAndSaveCredential()

	t.Logf("outbuf: %s", outbuf.String())
	// Test
	assert.NoError(err)
	assert.Contains(outbuf.String(), "Username:")
	assert.Contains(outbuf.String(), "Password:")
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
			credentialPath = tc
			machine := "aperture.section.io"
			username := "ada@lovelace.example"
			password := "s3cr3t"

			// Invoke
			err := WriteCredential(machine, username, password)

			// Test
			assert.NoError(err)
			info, err := os.Stat(credentialPath)
			assert.NoError(err)
			assert.Equal(info.Mode().Perm(), os.FileMode(0x180)) // 0600
		})
	}
}

func TestAPIAuthPromptRecordsCredential(t *testing.T) {
	assert := assert.New(t)

	// Setup
	credentialPath = newCredentialTempfile(t)

	var outbuf bytes.Buffer
	out = &outbuf

	text := "jane@section.example\ns3cr3t\n"
	c := exec.Command("echo", "-e", "'"+text+"'")
	tty, err := pty.Start(c)
	assert.NoError(err)
	in = tty
	defer func() { in = os.Stdin }()

	/*
		// Invoke
		ConsentGiven, err := ReadConsent()

		// Test
		assert.True(ConsentGiven)
		assert.NoError(err)
		consentFile, err := os.Open(consentPath)
		assert.NoError(err)
		contents, err := ioutil.ReadAll(consentFile)
		assert.NoError(err)
		var consent cliTrackingConsent
		err = json.Unmarshal(contents, &consent)
		assert.NoError(err)
		assert.True(consent.ConsentGiven)
	*/
}
func TestAPIAuthGetBasicAuthReturnsCredential(t *testing.T)               {}
func TestAPIAuthGetBasicAuthReturnsErrorIfCredentialInvalid(t *testing.T) {}
