package commands

import (
	"fmt"
	"strings"

	"github.com/section/sectionctl/api"
	"github.com/section/sectionctl/api/auth"
)

// LoginCmd handles authenticating the CLI against Section's API
type LoginCmd struct{}

// Run executes the command
func (c *LoginCmd) Run() (err error) {
	fmt.Printf("Setting up your authentication for %s...\n\n", api.PrefixURI.Host)

	u, p, err := auth.PromptForCredential(api.PrefixURI.Host)
	if err != nil {
		return fmt.Errorf("error when prompting for credential: %s", err)
	}

	err = auth.WriteCredential(api.PrefixURI.Host, u, p)
	if err != nil {
		return fmt.Errorf("unable to save credential: %s", err)
	}

	fmt.Print("Validating credentials...")
	_, err = api.CurrentUser()
	if err != nil {
		fmt.Println("error!")
		if strings.Contains(err.Error(), "401") {
			return fmt.Errorf("invalid credentials. Please try again")
		}
		return fmt.Errorf("could not fetch current user: %s", err)
	}
	fmt.Println("success!")

	return err
}
