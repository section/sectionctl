package auth

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/zalando/go-keyring"
)

func TestAPIAuthDetectsIfCredentialNotRecorded(t *testing.T) {
	assert := assert.New(t)

	assert.False(IsCredentialRecorded(KeyringService, t.Name()))
}

func TestAPIAuthPromptsForCredential(t *testing.T) {
	assert := assert.New(t)

	// Setup
	token := "s3cr3t"
	in := strings.NewReader(token + "\n")
	var out bytes.Buffer

	// Invoke
	to, err := PromptForCredential(in, &out)

	// Test
	assert.NoError(err)
	assert.Equal(token, to)
}

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
	keyring.MockInit()
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
