package commands

import (
	"errors"
	"fmt"
	"io"
	"os"
	"runtime"

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
	if api.Token != "" {
		err = credentials.Write(api.PrefixURI.Host, api.Token)
		if err != nil {
			if runtime.GOOS == "linux" || runtime.GOOS == "darwin" {
				fmt.Printf("Unable to write credential.\n\nPlease run this command, and add it to your ~/.bashrc\n\nexport SECTION_TOKEN=%s\n",api.Token)
				return nil
			} else if runtime.GOOS == "windows" {
				screenshot := "https://raw.githubusercontent.com/section/sectionctl/main/docs/section_token_control_panel.png"
				fmt.Printf("Unable to write credential.\n\nPlease execute the following, add it to your Powershell profile, or add it to your environment variables in control panel: \nWith Powershell:\n$env:SECTION_TOKEN=\"%s\"\n\nWith CMD:\nset SECTION_TOKEN=%s\n\nWith control panel:\n%s",api.Token,api.Token,screenshot)
				return nil
			}
			return fmt.Errorf("unable to write credential: %w", err)
		}
	} else {
		t, err := credentials.PromptAndWrite(c.In(), c.Out(), api.PrefixURI.Host)
		if err != nil {
			return fmt.Errorf("unable to prompt and write credentials: %w", err)
		}
		api.Token = t
	}
	fmt.Print("\nValidating credentials...")
	_, err = api.CurrentUser()
	if err != nil {
		fmt.Println("error!")
		if errors.Is(err, api.ErrAuthDenied) {
			return err
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
