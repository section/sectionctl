package analytics

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/tcnksm/go-gitconfig"
	"net/http"
	"time"
)

var (
	// HeapBaseURI is the location of Heap API endpoint
	HeapBaseURI = "https://heapanalytics.com"
	// HeapAppID identifies what Heap App events are recorded against
	HeapAppID = "4248790180"
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

// Submit submits an analytics event to Section
func Submit(e Event) (err error) {
	email, err := gitconfig.Email()
	if err != nil {
		return err
	}

	hte := heapTrackEvent{
		Identity:   email,
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
	resp, err := http.Post(uri, "application/json", bytes.NewBuffer(json))
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return errors.New("bad response")
	}
	return nil
}
