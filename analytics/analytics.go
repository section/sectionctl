package analytics

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/alecthomas/kong"
	"github.com/denisbrodbeck/machineid"
	"io"
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
	// out is the buffer to write feedback to users
	out io.Writer
	// in is the buffer to read feedback from users
	in io.Reader
)

func init() {
	// Set in + out for normal interactive use
	in = os.Stdin
	out = os.Stdout
}

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

	contents, err := ioutil.ReadAll(consentFile)
	if err != nil {
		return false
	}

	var consent cliTrackingConsent
	err = json.Unmarshal(contents, &consent)

	return err == nil
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
	err = json.Unmarshal(contents, &consent)
	if err != nil {
		return
	}

	ConsentGiven = consent.ConsentGiven
}

// Println formats using the default formats for its operands and writes to output.
// Output can be overridden for testing purposes by setting: analytics.out
// It returns the number of bytes written and any write error encountered.
func Println(a ...interface{}) (n int, err error) {
	return fmt.Fprintln(out, a...)
}

// Printf formats according to a format specifier and writes to standard output.
// Output can be overridden for testing purposes by setting: analytics.out
// It returns the number of bytes written and any write error encountered.
func Printf(format string, a ...interface{}) (n int, err error) {
	return fmt.Fprintf(out, format, a...)
}

// PromptForConsent interactively prompts the user for consent
func PromptForConsent() {
	Printf("👋 Hi!\n\n")
	Printf("Thanks for using the Section CLI.\n\n")
	Printf("We're still working out how to create the best experience for you.\n\n")
	Printf("We'd like to collect some anonymous info about how you use the CLI.\n\n")
	Printf("Are you OK with this? [y/N] ")

	reader := bufio.NewReader(in)
	text, err := reader.ReadString('\n')
	if err != nil {
		Println("Error: unable to read your response. Exiting.")
		os.Exit(2)
	}
	text = strings.Replace(text, "\n", "", -1) // convert CRLF to LF

	if strings.EqualFold(text, "y") {
		Printf("\nThank you!\n")
		ConsentGiven = true
	} else {
		ConsentGiven = false
		Printf("\nNo worries! We won't ask again.\n")
	}

	writeConsent()
}

// writeConsent writes the current consent state to a persistent file
func writeConsent() {
	c := cliTrackingConsent{ConsentGiven: ConsentGiven}
	json, err := json.Marshal(c)
	if err != nil {
		Println("Error: unable to record consent. Exiting.")
		os.Exit(2)
	}
	err = ioutil.WriteFile(consentPath, json, 0644)
	if err != nil {
		Println("Error: unable to record consent. Exiting.")
		os.Exit(2)
	}
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
	if !ConsentGiven {
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
	json, err := json.Marshal(ev)
	if err != nil {
		return err
	}

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
