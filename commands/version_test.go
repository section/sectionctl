package commands

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCommandsVersionHandlesVersionCheckTimeout(t *testing.T) {
	assert := assert.New(t)

	c := VersionCmd{}
	c.Run()

	assert.True(false)
}
