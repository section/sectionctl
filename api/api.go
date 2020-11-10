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
	// Username is the username for authenticating to the Section API
	Username string
	// Token is the token for authenticating to the Section API
	Token string
	// Debug toggles whether extra information is emitted from requests/responses
	Debug bool
)

// BaseURL returns a URL for building requests on
func BaseURL() (u url.URL) {
	u = *PrefixURI
	u.Path += "/api/v1"
	return u
}

func request(method string, u url.URL, body io.Reader, headers ...map[string][]string) (resp *http.Response, err error) {
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
	if len(headers) == 1 {
		for h, v := range headers[0] {
			req.Header[h] = v
		}
	}

	if Username == "" || Token == "" {
		Username, Token, err = auth.GetCredential(u.Host)
		if err != nil {
			return resp, err
		}
	}
	req.SetBasicAuth(Username, Token)

	if Debug {
		fmt.Println("[DEBUG] Request URL:", req.URL)
		for k, vs := range req.Header {
			for _, v := range vs {
				fmt.Printf("[DEBUG] Header: %s: %v\n", k, v)
			}
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
	if len(resp.Header["Aperture-Tx-Id"]) > 0 {
		return fmt.Errorf("request failed with status %s and transaction ID %s", resp.Status, resp.Header["Aperture-Tx-Id"][0])
	}
	return fmt.Errorf("request failed with status %s", resp.Status)
}
