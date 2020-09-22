package commands

import (
	"fmt"

	"github.com/section/section-cli/api/auth"
)

// LoginCmd handles authenticating the CLI against Section's API
type LoginCmd struct{}

// Run executes the command
func (c *LoginCmd) Run() (err error) {
	fmt.Printf("Setting up your authentication for the Section API...\n\n")

	m, u, p, err := auth.PromptForCredential()
	if err != nil {
		return fmt.Errorf("error when prompting for credential: %s", err)
	}

	err = auth.WriteCredential(m, u, p)
	if err != nil {
		return fmt.Errorf("unable to save credential: %s", err)
	}

	fmt.Println("Success!")

	return err
}
