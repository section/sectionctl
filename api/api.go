package api

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"runtime"
	"time"

	"github.com/section/section-cli/api/auth"
	"github.com/section/section-cli/version"
)

var (
	// PrefixURI is the root of the Section API
	PrefixURI = "https://aperture.section.io"
	timeout   = 20 * time.Second
)

// BaseURL returns a URL for building requests on
func BaseURL() (*url.URL, error) {
	return url.Parse(PrefixURI + "/api/v1")
}

func request(method string, url string, body io.Reader) (resp *http.Response, err error) {
	client := &http.Client{
		Timeout: timeout,
	}

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return resp, err
	}

	ua := fmt.Sprintf("section-cli (%s; %s-%s)", version.Version, runtime.GOARCH, runtime.GOOS)
	req.Header.Set("User-Agent", ua)

	username, password, err := auth.GetBasicAuth()
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
