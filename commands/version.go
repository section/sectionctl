package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"runtime"
	"time"

	"github.com/fatih/color"
	hversion "github.com/hashicorp/go-version"
	"github.com/section/sectionctl/version"
)

// VersionCmd handles authenticating the CLI against Section's API
type VersionCmd struct {
	in               io.Reader
	out              io.Writer
	LatestReleaseURL *url.URL      `hidden default:"https://api.github.com/repos/section/sectionctl/releases/latest"`
	Timeout          time.Duration `hidden default:"5s"`
}

// Run executes the `login` command
func (c *VersionCmd) Run() (err error) {
	reply := make(chan string, 1)
	errs := make(chan error, 1)
	go c.checkVersion(reply, errs)

	fmt.Fprintf(c.Out(), "%s\n", c.String())

	if version.Version == "dev" {
		return err
	}

	var v string
	select {
	case v = <-reply:
	case err = <-errs:
		return err
	}

	current, err := hversion.NewVersion(version.Version)
	if err != nil {
		return err
	}
	latest, err := hversion.NewVersion(v)
	if err != nil {
		return err
	}

	if current.LessThan(latest) {
		green := color.New(color.FgGreen)
		green.Fprintf(c.Out(), "\nA new version of sectionctl is available: %s\n", v)
		fmt.Fprintf(c.Out(), "\nDownload at https://github.com/section/sectionctl/releases/%s\n", v)
	}

	return err
}

func (c *VersionCmd) checkVersion(latest chan string, errs chan error) {
	ctx, cancel := context.WithTimeout(context.Background(), c.Timeout)
	defer cancel()
	if c.LatestReleaseURL == nil {
		errs <- fmt.Errorf("no release URL specified")
		return
	}
	req, err := http.NewRequestWithContext(ctx, "GET", c.LatestReleaseURL.String(), nil)
	if err != nil {
		errs <- fmt.Errorf("unable to make request: %w", err)
		return
	}
	ua := fmt.Sprintf("sectionctl (%s; %s-%s)", version.Version, runtime.GOARCH, runtime.GOOS)
	req.Header.Set("User-Agent", ua)

	var client http.Client
	resp, err := client.Do(req)
	if err != nil {
		errs <- fmt.Errorf("unable to perform request: %w", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		errs <- fmt.Errorf("bad response from GitHub. Please try again later")
		return
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		errs <- fmt.Errorf("unable to read response: %w", err)
		return
	}

	var latestResp struct {
		TagName string `json:"tag_name"`
	}
	err = json.Unmarshal(body, &latestResp)
	if err != nil {
		errs <- fmt.Errorf("unable to decode response: %w", err)
		return
	}
	latest <- latestResp.TagName
}

func (c VersionCmd) String() string {
	return fmt.Sprintf("%s (%s-%s)", version.Version, runtime.GOOS, runtime.GOARCH)
}

// In returns the input to read from
func (c *VersionCmd) In() io.Reader {
	if c.in != nil {
		return c.in
	}
	return os.Stdin
}

// Out returns the output to write to
func (c *VersionCmd) Out() io.Writer {
	if c.out != nil {
		return c.out
	}
	return os.Stdout
}
