package commands

import (
	"fmt"
	"net/http"
	"runtime"
)

// Version is the version number of the running app
var Version = "0.0.1"

// VersionCheckEndpoint is where the CLI should check for version information
var VersionCheckEndpoint = "https://www.section.io/cli"

// VersionCmd handles versioning for the Section CLI
type VersionCmd struct{}

// Run executes the `login` command
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
