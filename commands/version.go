package commands

import (
	"fmt"
	"runtime"
)

// Version is the version number of the running app
var Version = "0.0.1"

// VersionCmd handles versioning for the Section CLI
type VersionCmd struct{}

// Run executes the `login` command
func (a *VersionCmd) Run() (err error) {
	fmt.Printf("%s (%s-%s)\n", Version, runtime.GOOS, runtime.GOARCH)
	return err
}
