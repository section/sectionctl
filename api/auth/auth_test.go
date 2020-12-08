package auth

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAPIAuthDetectsIfCredentialNotRecorded(t *testing.T) {
	assert := assert.New(t)

	assert.False(IsCredentialRecorded(KeyringService, t.Name()))
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

func TestAPIAuthGetCredentialReturnsErrorIfNone(t *testing.T) {
	assert := assert.New(t)

	// Invoke
	_, err := GetCredential(t.Name())

	// Test
	assert.Error(err)
	assert.False(IsCredentialRecorded(KeyringService, t.Name()))
}

func TestAPIAuthCanReadWrittenCredentials(t *testing.T) {
	assert := assert.New(t)

	// Setup
	endpoint := "127.0.0.1:8080"
	token := "s3cr3t"

	// Invoke
	err := WriteCredential(endpoint, token)
	assert.NoError(err)

	// Test
	to, err := GetCredential(endpoint)
	assert.NoError(err)
	assert.Equal(token, to)
}
