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
				fmt.Printf("Please execute the following or add it to your ~/.bashrc : \nexport SECTION_TOKEN=%s\n",api.Token)
			}
			return nil
		}
	} else {
		t, err := credentials.PromptAndWrite(c.In(), c.Out(), api.PrefixURI.Host)
		if err != nil {
			return fmt.Errorf("unable to prompt and write credentials: %w", err)
		}
		api.Token = t
	}
	if runtime.GOOS == "windows" {
		fmt.Printf("Please execute the following, add it to your Powershell profile, or add it to your environment variables in control panel: \nWith Powershell:\n$env:SECTION_TOKEN=\"%s\"\n\nWith CMD:\nset SECTION_TOKEN=%s\n\nWith control panel:\nhttps://raw.githubusercontent.com/section/sectionctl/main/docs/section_token_control_panel.png",api.Token,api.Token)
		return nil
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
