package analytics

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/alecthomas/kong"
	"github.com/denisbrodbeck/machineid"
	"io/ioutil"
	"net/http"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"time"
)

var (
	// HeapBaseURI is the location of Heap API endpoint
	HeapBaseURI = "https://heapanalytics.com"
	// HeapAppID identifies what Heap App events are recorded against
	HeapAppID = "4248790180"
	// consentPath is the path on disk to where analytics consent is recorded
	consentPath string
	// ConsentGiven records whether tracking consent has been given by the user
	ConsentGiven bool
)

// Private type for submitting server side events to Heap
// https://developers.heap.io/reference#track-1
type heapTrack struct {
	AppID  string           `json:"app_id"`
	Events []heapTrackEvent `json:"events"`
}

// Private type for submitting server side events to Heap
// https://developers.heap.io/reference#track-1
type heapTrackEvent struct {
	Identity   string            `json:"identity"`
	Timestamp  time.Time         `json:"timestamp"`
	Event      string            `json:"event"`
	Properties map[string]string `json:"properties"`
}

// Event records an interaction with the Section CLI
type Event struct {
	Name       string
	Properties map[string]string
}

// Identity tries to determine the identity of the machine the cli is being run on
func identity() (id string) {
	id, err := machineid.ProtectedID("section-cli")
	if err == nil {
		return id
	}
	return "unknown"
}

// LogInvoke logs an invocation of the cli
func LogInvoke(ctx *kong.Context) {
	props := map[string]string{
		"Subcommand": ctx.Command(),
		"Args":       strings.Join(ctx.Args, " "),
	}
	if ctx.Error != nil {
		props["Error"] = ctx.Error.Error()
	}
	e := Event{
		Name:       "CLI invoked",
		Properties: props,
	}
	err := Submit(e)
	if err != nil {
		fmt.Println("Warning: Unable to submit analytics – continuing anyway.")
	}
}

type cliTrackingConsent struct {
	ConsentGiven bool `json:"consent_given"`
}

// IsConsentRecorded returns if valid consent has been recorded for tracking
func IsConsentRecorded() (rec bool) {
	if len(consentPath) == 0 {
		usr, err := user.Current()
		if err != nil {
			fmt.Println("user")
			return false
		}
		consentPath = filepath.Join(usr.HomeDir, ".config", "section", "cli_tracking_consent")
	}
	if _, err := os.Stat(consentPath); err != nil {
		return false
	}

	consentFile, err := os.Open(consentPath)
	if err != nil {
		return false
	}
	defer consentFile.Close()

	var consent cliTrackingConsent
	contents, err := ioutil.ReadAll(consentFile)
	if err != nil {
		return false
	}
	err = json.Unmarshal(contents, &consent)
	if err != nil {
		return false
	}

	return true
}

// ReadConsent finds if consent has been given
func ReadConsent() {
	if !IsConsentRecorded() {
		PromptForConsent()
	}

	if _, err := os.Stat(consentPath); err != nil {
		return
	}

	consentFile, err := os.Open(consentPath)
	if err != nil {
		return
	}
	defer consentFile.Close()

	var consent cliTrackingConsent
	contents, err := ioutil.ReadAll(consentFile)
	if err != nil {
		return
	}
	json.Unmarshal(contents, &consent)

	ConsentGiven = consent.ConsentGiven
}

// PromptForConsent interactively prompts the user for consent
func PromptForConsent() {
}

// Submit submits an analytics event to Section
//
// Behavior is determined by consent:
//
// if consent not given {
// 	prompt for consent
// }
// if consent == true {
// 	submit analytics
// }
func Submit(e Event) (err error) {
	ReadConsent()
	if ConsentGiven == false {
		return err
	}

	hte := heapTrackEvent{
		Identity:   identity(),
		Event:      e.Name,
		Timestamp:  time.Now(),
		Properties: e.Properties,
	}
	ev := heapTrack{
		AppID:  HeapAppID,
		Events: []heapTrackEvent{hte},
	}
	json, _ := json.Marshal(ev)

	uri := fmt.Sprintf("%s/api/track", HeapBaseURI)
	c := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest(http.MethodPost, uri, bytes.NewBuffer(json))
	req.Header.Add("Content-Type", "application/json")
	if err != nil {
		return err
	}
	resp, err := c.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return errors.New("bad response")
	}
	return nil
}
