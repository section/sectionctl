package commands

import (
	"fmt"
	"net/http"
	"runtime"
)

var (
	// Version is the version number of the running app
	Version = "0.0.1"
	// VersionCheckEndpoint is where the CLI should check for version information
	VersionCheckEndpoint = "https://www.section.io/cli/current_version"
)

// VersionCmd handles versioning for the Section CLI
type VersionCmd struct{}

// Run executes the command
func (c *VersionCmd) Run() (err error) {
	fmt.Printf("%s (%s-%s)\n", Version, runtime.GOOS, runtime.GOARCH)
	c.checkVersion()
	return err
}

func (c *VersionCmd) checkVersion() {
	fmt.Println(VersionCheckEndpoint)
	resp, err := http.Get(VersionCheckEndpoint)
	if err != nil {
		fmt.Printf("Error: unable to check version: %s", err)
	}
	fmt.Printf("%+v\n", resp)
}
