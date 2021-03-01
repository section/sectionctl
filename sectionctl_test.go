package main

import (
	"io/ioutil"
	"net/url"
	"path/filepath"
	"testing"

	"github.com/alecthomas/kong"
	"github.com/stretchr/testify/assert"
)

func TestBootstrapSetsUpDebugLogFile(t *testing.T) {
	assert := assert.New(t)

	// Setup
	u, err := url.Parse("https://aperture.section.io")
	assert.NoError(err)
	d, err := ioutil.TempDir("", "sectionctl")
	assert.NoError(err)

	var cli = CLI{Debug: true, SectionToken: "s3cr3t", SectionAPIPrefix: u, DebugFileDir: d}
	var ctx kong.Context

	// Invoke
	bootstrap(cli, &ctx)

	// Test
	m, err := filepath.Glob(filepath.Join(d, "*.log"))
	assert.NoError(err)
	assert.Greater(len(m), 0)
}
