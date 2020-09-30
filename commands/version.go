package commands

import (
	"fmt"
	"runtime"

	"github.com/section/sectionctl/version"
)

// VersionCmd handles authenticating the CLI against Section's API
type VersionCmd struct{}

// Run executes the `login` command
func (c *VersionCmd) Run() (err error) {
	fmt.Printf("%s (%s-%s)\n", version.Version, runtime.GOOS, runtime.GOARCH)
	return err
}
