package api

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"runtime"
	"time"

	"github.com/section/sectionctl/version"
)

var (
	// PrefixURI is the root of the Section API
	PrefixURI = &url.URL{Scheme: "https", Host: "aperture.section.io"}
	// Timeout specifies a time limit for requests made by the API client
	Timeout = 20 * time.Second
	// Token is the token for authenticating to the Section API
	Token string

	// ErrAuthDenied represents all authentication and authorization errors
	ErrAuthDenied = errors.New("denied")
	// ErrStatusUnauthorized (401) indicates that the request has not been applied because it lacks valid authentication credentials for the target resource.
	ErrStatusUnauthorized = fmt.Errorf("check your token? API request is unauthorized: %w", ErrAuthDenied)
	// ErrStatusForbidden (403) indicates that the server understood the request but refuses to authorize it.
	ErrStatusForbidden = fmt.Errorf("check your token? API request is forbidden: %w", ErrAuthDenied)

	// client is the HTTP client used across requests
	client http.Client
)

// BaseURL returns a URL for building requests on
func BaseURL() (u url.URL) {
	u = *PrefixURI
	u.Path += "/api/v1"
	return u
}

// request does the heavy lifting of making requests to the Section API.
//
// You can pass 0 or more headers, and keys in the later headers will override earlier passed headers.
func request(ctx context.Context, method string, u url.URL, body io.Reader, headers ...http.Header) (resp *http.Response, err error) {
	req, err := http.NewRequestWithContext(ctx, method, u.String(), body)
	if err != nil {
		return resp, err
	}

	ua := fmt.Sprintf("sectionctl (%s; %s-%s)", version.Version, runtime.GOARCH, runtime.GOOS)
	req.Header.Set("User-Agent", ua)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")
	for i := range headers {
		for h, v := range headers[i] {
			req.Header[h] = v
		}
	}

	req.Header.Add("section-token", Token)

	log.Println("[DEBUG] Request URL:", method, req.URL)
	for k, vs := range req.Header {
		for _, v := range vs {
			log.Printf("[DEBUG] Header: %s: %v\n", k, v)
		}
	}
	resp, err = client.Do(req)
	if err != nil {
		return resp, err
	}
	return resp, err
}

// prettyTxIDError creates a support friendly error message with an transaction ID
func prettyTxIDError(resp *http.Response) error {
	var txIDhelper string
	if len(resp.Header["Aperture-Tx-Id"]) > 0 {
		txIDhelper = fmt.Sprintf(" and Section Transaction ID %s", resp.Header["Aperture-Tx-Id"][0])
	}

	switch {
	case resp.StatusCode == http.StatusTooManyRequests:
		return fmt.Errorf("status 429%s: the number of requests have exceeded the maximum allowed for this time period. Please wait a few minutes and try again", txIDhelper)
	default:
		return fmt.Errorf("request failed with status %s%s", resp.Status, txIDhelper)

	}
}
