package commands

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"runtime"
)

var (
	// Version is the version number of the running app
	Version = "0.0.1"
	// VersionCheckURL is where the CLI should check for version information
	VersionCheckURL = "https://www.section.io/cli/last_version"
)

// VersionCmd handles versioning for the Section CLI
type VersionCmd struct{}

// Run executes the command
func (c *VersionCmd) Run() (err error) {
	fmt.Printf("%s (%s-%s)\n", Version, runtime.GOOS, runtime.GOARCH)
	v, err := c.checkVersion()
	if err != nil {
		fmt.Printf("Error: unable to check version: %s", err)
		os.Exit(2)
	}
	fmt.Printf("Latest version: %s\n", v)
	return err
}

// versionCheck is the response
type versionCheck struct {
	LatestVersion string `json:"latest_version"`
}

func (c *VersionCmd) checkVersion() (version string, err error) {
	resp, err := http.Get(VersionCheckURL)
	if err != nil {
		return version, err
	}

	if resp.StatusCode != http.StatusOK {
		return version, errors.New("bad response when checking version")
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return version, err
	}
	var vc versionCheck
	err = json.Unmarshal(body, &vc)

	return vc.LatestVersion, err
}
