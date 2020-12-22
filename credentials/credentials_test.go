package credentials

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/zalando/go-keyring"
)

func TestCredentialsDetectsIfCredentialNotRecorded(t *testing.T) {
	assert := assert.New(t)
	keyring.MockInit()
	assert.False(IsCredentialRecorded(KeyringService, t.Name()))
}

func TestCredentialsPromptsForCredential(t *testing.T) {
	assert := assert.New(t)

	// Setup
	token := "s3cr3t"
	in := strings.NewReader(token + "\n")
	var out bytes.Buffer

	// Invoke
	to, err := Prompt(in, &out)

	// Test
	assert.NoError(err)
	assert.Equal(token, to)
}

func TestCredentialsGetCredentialReturnsErrorIfNone(t *testing.T) {
	assert := assert.New(t)
	keyring.MockInit()

	// Invoke
	_, err := Read(t.Name())

	// Test
	assert.Error(err)
	assert.False(IsCredentialRecorded(KeyringService, t.Name()))
}

func TestCredentialsCanReadWrittenCredentials(t *testing.T) {
	assert := assert.New(t)

	// Setup
	keyring.MockInit()
	endpoint := "127.0.0.1:8080"
	token := "s3cr3t"

	// Invoke
	err := Write(endpoint, token)
	assert.NoError(err)

	// Test
	to, err := Read(endpoint)
	assert.NoError(err)
	assert.Equal(token, to)
}
