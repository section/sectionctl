package commands

import (
	"fmt"

	"github.com/alecthomas/kong"
	"github.com/section/sectionctl/api"
	"github.com/section/sectionctl/credentials"
)

// LogoutCmd handles revoking previously set up authentication
type LogoutCmd struct{}

// Run executes the command
func (c *LogoutCmd) Run(cli *CLI, ctx *kong.Context,logWriters *LogWriters) (err error) {
	s := NewSpinner(cli, fmt.Sprintf("Revoking your authentication for %s", api.PrefixURI.Host),logWriters)
	s.Start()
	err = credentials.Delete(api.PrefixURI.Host)
	s.Stop()
	return err
}
