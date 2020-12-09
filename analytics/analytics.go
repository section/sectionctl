package analytics

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"
	"time"

	"github.com/alecthomas/kong"
	"github.com/denisbrodbeck/machineid"
	"github.com/section/sectionctl/version"
)

var (
	// HeapBaseURI is the location of Heap API endpoint
	HeapBaseURI = "https://heapanalytics.com"
	// HeapAppID identifies what Heap App events are recorded against
	HeapAppID = "4248790180" // development id, overridden during `make build`
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
	id, err := machineid.ProtectedID("sectionctl")
	if err == nil {
		return id
	}
	return "unknown"
}

// AsyncLogInvoke logs an invocation of the cli
func AsyncLogInvoke(ctx *kong.Context) {
	if ctx.Command() == "analytics" {
		return
	}

	ConsentGiven, err := ReadConsent(os.Stdin, os.Stdout)
	if err != nil || !ConsentGiven {
		return
	}

	exe, err := os.Executable()
	if err != nil {
		log.Printf("[WARN] Unable to submit analytics: %s", err)
		return
	}

	props := map[string]string{
		"Subcommand": ctx.Command(),
		"Args":       strings.Join(ctx.Args, " "),
		"Version":    version.Version,
	}
	if ctx.Error != nil {
		props["Error"] = ctx.Error.Error()
	}
	e := Event{
		Name:       "CLI invoked",
		Properties: props,
	}

	j, err := json.Marshal(e)

	cmd := exec.Command(exe, "analytics", "--event", string(j))
	err = cmd.Start()
	if err != nil {
		log.Printf("[WARN] Unable to submit analytics: %s", err)
	}
}

// AsyncLogError logs an invocation of the cli
func AsyncLogError(ctx *kong.Context, uerr error) {
	if ctx.Command() == "analytics" {
		return
	}

	ConsentGiven, err := ReadConsent(os.Stdin, os.Stdout)
	if err != nil || !ConsentGiven {
		return
	}

	exe, err := os.Executable()
	if err != nil {
		log.Printf("[WARN] Unable to submit analytics: %s", err)
		return
	}

	e := Event{
		Name: "sectionctl error",
		Properties: map[string]string{
			"Subcommand": ctx.Command(),
			"Args":       strings.Join(ctx.Args, " "),
			"Version":    version.Version,
			"Error":      fmt.Sprintf("%s", uerr),
		},
	}

	j, err := json.Marshal(e)

	cmd := exec.Command(exe, "analytics", "--event", string(j))
	err = cmd.Start()
	if err != nil {
		log.Printf("[WARN] Unable to submit analytics: %s", err)
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

// ReadConsent finds if consent has been given, either from file or by prompt.
func ReadConsent(in io.Reader, out io.Writer) (c bool, err error) {
	if !IsConsentRecorded() {
		c, err = PromptForConsent(in, out)

		err = WriteConsent(c)
		if err != nil {
			return c, fmt.Errorf("unable to record consent: %s", err)
		}
	}

	if _, err := os.Stat(consentPath); err != nil {
		return c, err
	}

	consentFile, err := os.Open(consentPath)
	if err != nil {
		return c, err
	}
	defer consentFile.Close()

	var consent cliTrackingConsent
	contents, err := ioutil.ReadAll(consentFile)
	if err != nil {
		return c, err
	}
	err = json.Unmarshal(contents, &consent)
	if err != nil {
		return c, err
	}

	return consent.ConsentGiven, err
}

// PromptForConsent interactively prompts the user for consent
func PromptForConsent(in io.Reader, out io.Writer) (c bool, err error) {
	fmt.Fprintf(out, "👋 Hi!\n\n")
	fmt.Fprintf(out, "Thanks for using the Section CLI.\n\n")
	fmt.Fprintf(out, "We're still working out how to create the best experience for you.\n\n")
	fmt.Fprintf(out, "We'd like to collect some anonymous info about how you use the CLI.\n\n")
	fmt.Fprintf(out, "Are you OK with this? [y/N] ")

	reader := bufio.NewReader(in)
	text, err := reader.ReadString('\n')
	if err != nil {
		return c, fmt.Errorf("unable to read your response: %s", err)
	}
	text = strings.Replace(text, "\n", "", -1)
	text = strings.Replace(text, "\r", "", -1) // convert CRLF to LF

	if strings.EqualFold(text, "y") {
		fmt.Fprintf(out, "\nThank you!\n")
		return true, err
	}
	fmt.Fprint(out, "\nNo worries! We won't ask again.\n")
	return false, err
}

// WriteConsent writes the current consent state to a persistent file
func WriteConsent(consent bool) (err error) {
	c := cliTrackingConsent{ConsentGiven: consent}
	json, err := json.Marshal(c)
	if err != nil {
		return err
	}
	consentPathBasedir := filepath.Dir(consentPath)
	err = os.MkdirAll(consentPathBasedir, os.ModeDir+0700)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(consentPath, json, 0644)
	if err != nil {
		return err
	}
	return err
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
	ConsentGiven, err = ReadConsent(os.Stdin, os.Stdout)
	if err != nil || !ConsentGiven {
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
