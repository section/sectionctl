package analytics

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func newConsentTempfile(t *testing.T) string {
	pattern := "sectionctl-analytics-consent-" + strings.ReplaceAll(t.Name(), "/", "_")
	file, err := ioutil.TempFile("", pattern)
	if err != nil {
		t.FailNow()
	}
	return file.Name()
}

func newConsentTempdir(t *testing.T) string {
	pattern := "sectionctl-analytics-consent-" + strings.ReplaceAll(t.Name(), "/", "_")
	dir, err := ioutil.TempDir("", pattern)
	if err != nil {
		t.FailNow()
	}
	return dir
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
	var inbuf bytes.Buffer
	inbuf.Write([]byte("\n"))

	// Invoke
	ConsentGiven, err := ReadConsent(&inbuf, &outbuf)

	// Test
	assert.False(ConsentGiven)
	assert.NoError(err)
	assert.Contains(outbuf.String(), "[y/N]")
}

func TestConsentPromptForConsent(t *testing.T) {
	assert := assert.New(t)

	var testCases = []struct {
		input  string
		retval bool
	}{
		{"y\n", true},
		{"n\n", false},
	}
	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			// Setup
			var outbuf bytes.Buffer
			var inbuf bytes.Buffer
			inbuf.Write([]byte(tc.input))

			// Invoke
			c, err := PromptForConsent(&inbuf, &outbuf)

			// Test
			assert.NoError(err)
			assert.Equal(c, tc.retval)
		})
	}
}

func TestConsentPromptRecordsConsent(t *testing.T) {
	assert := assert.New(t)

	// Setup
	consentPath = newConsentTempfile(t)
	assert.False(IsConsentRecorded())

	var outbuf bytes.Buffer
	var inbuf bytes.Buffer
	inbuf.Write([]byte("y\n"))

	// Invoke
	ConsentGiven, err := ReadConsent(&inbuf, &outbuf)

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
			var inbuf bytes.Buffer
			inbuf.Write([]byte(tc))

			// Invoke
			ConsentGiven, err := ReadConsent(&inbuf, &outbuf)

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

func TestConsentPromptHandlesNewlines(t *testing.T) {
	assert := assert.New(t)

	testCases := []string{
		"y\n",
		"y\r\n",
		"Y\n",
		"Y\r\n",
	}

	for _, tc := range testCases {
		t.Run(tc, func(t *testing.T) {
			// Setup
			consentPath = newConsentTempfile(t)

			var outbuf bytes.Buffer
			var inbuf bytes.Buffer
			inbuf.Write([]byte(tc))

			// Invoke
			ConsentGiven, err := ReadConsent(&inbuf, &outbuf)

			// Test
			assert.True(ConsentGiven)
			assert.NoError(err)
			assert.Contains(outbuf.String(), "[y/N]")
		})
	}
}

func TestConsentWriteConsentCreatesPathIfItDoesNotExist(t *testing.T) {
	assert := assert.New(t)
	// Setup
	consentPath = filepath.Join(newConsentTempdir(t), "does", "not", "exist")

	err := WriteConsent(ConsentGiven)
	assert.NoError(err)

	info, err := os.Stat(consentPath)
	assert.NoError(err)
	assert.False(info.IsDir())

	info, err = os.Stat(filepath.Dir(consentPath))
	assert.NoError(err)
	assert.True(info.IsDir())
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
