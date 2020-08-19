package commands

import (
	"fmt"
	"runtime"
)

// VersionCmd handles authenticating the CLI against Section's API
type VersionCmd struct{}

// Run executes the `login` command
func (c *VersionCmd) Run() (err error) {
	fmt.Printf("%s (%s-%s)\n", "0.0.1", runtime.GOOS, runtime.GOARCH)
	return err
}
