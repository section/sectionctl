package analytics

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func newConsentTempfile(t *testing.T) string {
	pattern := "section-cli-analytics-consent-" + strings.ReplaceAll(t.Name(), "/", "_")
	file, err := ioutil.TempFile("", pattern)
	if err != nil {
		t.FailNow()
	}
	return file.Name()
}

func TestConsentDetectsIfConsentNotRecorded(t *testing.T) {
	assert := assert.New(t)

	// Setup
	consentPath = newConsentTempfile(t)

	// Test
	assert.False(IsConsentRecorded())
}

func TestConsentPromptsForConsentIfConsentNotRecorded(t *testing.T) {
	assert := assert.New(t)

	// Setup
	consentPath = newConsentTempfile(t)

	var outbuf bytes.Buffer
	out = &outbuf
	var inbuf bytes.Buffer
	inbuf.Write([]byte("\n"))
	in = &inbuf

	// Invoke
	ConsentGiven, err := ReadConsent()

	// Test
	assert.False(ConsentGiven)
	assert.NoError(err)
	assert.Contains(outbuf.String(), "[y/N]")
}

func TestConsentPromptRecordsConsent(t *testing.T) {
	assert := assert.New(t)

	// Setup
	consentPath = newConsentTempfile(t)

	var outbuf bytes.Buffer
	out = &outbuf

	var inbuf bytes.Buffer
	inbuf.Write([]byte("y\n"))
	in = &inbuf

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
}

func TestConsentPromptDefaultsToFalse(t *testing.T) {
	assert := assert.New(t)

	testCases := []string{
		"\n",
		"n\n",
		"N\n",
		"OESNAOSENAO\n",
		"😈\n",
	}
	for _, tc := range testCases {
		t.Run(tc, func(t *testing.T) {
			//			t.Parallel()
			// Setup
			consentPath = newConsentTempfile(t)

			var outbuf bytes.Buffer
			out = &outbuf

			var inbuf bytes.Buffer
			inbuf.Write([]byte(tc))
			in = &inbuf

			// Invoke
			ConsentGiven, err := ReadConsent()

			// Test
			assert.False(ConsentGiven)
			assert.NoError(err)
			consentFile, err := os.Open(consentPath)
			assert.NoError(err)
			contents, err := ioutil.ReadAll(consentFile)
			assert.NoError(err)
			var consent cliTrackingConsent
			err = json.Unmarshal(contents, &consent)
			assert.NoError(err)
			assert.False(consent.ConsentGiven)
		})
	}
}

func TestConsentSubmitNoopsIfNoConsent(t *testing.T) {
	assert := assert.New(t)
	var called bool
	consentPath = newConsentTempfile(t)
	ConsentGiven = false
	err := WriteConsent(ConsentGiven)
	assert.NoError(err)

	// Setup
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true

		body, _ := ioutil.ReadAll(r.Body)
		t.Logf("%s", body)
	}))
	HeapBaseURI = ts.URL

	// Invoke
	e := Event{
		Name: "CLI invoked",
		Properties: map[string]string{
			"Args":       "apps list",
			"Subcommand": "apps list",
			"Errors":     "",
		},
	}
	err = Submit(e)

	// Test
	assert.NoError(err)
	assert.False(called)
}
