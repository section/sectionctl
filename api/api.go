package api

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os/user"
	"path/filepath"
	"runtime"
	"time"

	"github.com/jdxcode/netrc"
	"github.com/section/section-cli/version"

	"github.com/section/section-cli/api/auth"
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

func getBasicAuth() (u, p string, err error) {
	usr, err := user.Current()
	if err != nil {
		panic(err)
	}
	n, err := netrc.Parse(filepath.Join(usr.HomeDir, ".netrc"))
	if err != nil {
		return u, p, err
	}
	u = n.Machine("aperture.section.io").Get("login")
	p = n.Machine("aperture.section.io").Get("password")
	return u, p, err
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
