package commands

import (
	"context"
	"fmt"

	"github.com/section/sectionctl/api"
	"github.com/section/sectionctl/credentials"
)

// LogoutCmd handles revoking previously set up authentication
type LogoutCmd struct{}

// Run executes the command
func (c *LogoutCmd) Run(ctx context.Context) (err error) {
	s := NewSpinner(ctx, fmt.Sprintf("Revoking your authentication for %s", api.PrefixURI.Host))
	s.Start()
	err = credentials.Delete(api.PrefixURI.Host)
	s.Stop()
	return err
}
