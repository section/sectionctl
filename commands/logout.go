package commands

import (
	"fmt"

	"github.com/section/sectionctl/api"
	"github.com/section/sectionctl/credentials"
)

// LogoutCmd handles revoking previously set up authentication
type LogoutCmd struct{}

// Run executes the command
func (c *LogoutCmd) Run(logWriters *LogWriters) (err error) {
	s := NewSpinner(fmt.Sprintf("Revoking your authentication for %s", api.PrefixURI.Host), logWriters)
	s.Start()
	err = credentials.Delete(api.PrefixURI.Host)
	s.Stop()
	return err
}
