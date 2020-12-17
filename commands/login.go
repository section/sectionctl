package commands

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/section/sectionctl/api"
	"github.com/section/sectionctl/credentials"
)

// LoginCmd handles authenticating the CLI against Section's API
type LoginCmd struct {
	in  io.Reader
	out io.Writer
}

// Run executes the command
func (c *LoginCmd) Run() (err error) {
	fmt.Printf("Setting up your authentication for %s...\n\n", api.PrefixURI.Host)

	t, err := credentials.Prompt(c.In(), c.Out())
	if err != nil {
		return fmt.Errorf("error when prompting for credential: %s", err)
	}

	err = credentials.Write(api.PrefixURI.Host, t)
	if err != nil {
		return fmt.Errorf("unable to save credential: %s", err)
	}

	fmt.Print("\nValidating credentials...")
	_, err = api.CurrentUser()
	if err != nil {
		fmt.Println("error!")
		if strings.Contains(err.Error(), `with status "4`) {
			return fmt.Errorf("invalid credentials. Please try again")
		}
		return fmt.Errorf("could not fetch current user: %w", err)
	}
	fmt.Println("success!")

	return err
}

// In returns the input to read from
func (c *LoginCmd) In() io.Reader {
	if c.in != nil {
		return c.in
	}
	return os.Stdin
}

// Out returns the output to write to
func (c *LoginCmd) Out() io.Writer {
	if c.out != nil {
		return c.out
	}
	return os.Stdout
}
