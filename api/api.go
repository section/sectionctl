package api

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"runtime"
	"time"

	"github.com/section/sectionctl/api/auth"
	"github.com/section/sectionctl/version"
)

var (
	// PrefixURI is the root of the Section API
	PrefixURI = &url.URL{Scheme: "https", Host: "aperture.section.io"}
	timeout   = 20 * time.Second
)

// BaseURL returns a URL for building requests on
func BaseURL() (u url.URL) {
	u = *PrefixURI
	u.Path += "/api/v1"
	return u
}

func request(method string, u url.URL, body io.Reader) (resp *http.Response, err error) {
	client := &http.Client{
		Timeout: timeout,
	}

	req, err := http.NewRequest(method, u.String(), body)
	if err != nil {
		return resp, err
	}

	ua := fmt.Sprintf("sectionctl (%s; %s-%s)", version.Version, runtime.GOARCH, runtime.GOOS)
	req.Header.Set("User-Agent", ua)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")

	username, password, err := auth.GetCredential(u.Host)
	if err != nil {
		return resp, err
	}
	req.SetBasicAuth(username, password)

	resp, err = client.Do(req)
	if err != nil {
		return resp, err
	}
	return resp, err
}
