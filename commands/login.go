package commands

import (
	"errors"
	"fmt"
	"io"
	"os"
	"runtime"

	"github.com/rs/zerolog/log"

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
	screenshot := "https://raw.githubusercontent.com/section/sectionctl/main/docs/section_token_control_panel.png"
	windowsStr := fmt.Sprintf("Unable to write credential.\n\nPlease execute the following, add it to your Powershell profile, or add it to your environment variables in control panel: \nWith Powershell:\n$env:SECTION_TOKEN=\"%s\"\n\nWith CMD:\nset SECTION_TOKEN=%s\n\nWith control panel:\n%s", api.Token, api.Token, screenshot)
	linuxStr := fmt.Sprintf("Unable to write credential.\n\nPlease run this command, and add it to your ~/.bashrc (you do not need to run sectionctl login again)\n\nexport SECTION_TOKEN=%s", api.Token)
	if api.Token != "" {
		err = credentials.Write(api.PrefixURI.Host, api.Token)
		if err != nil {
			if runtime.GOOS == "windows" {
				fmt.Print(windowsStr)
				return nil
			}
			fmt.Printf("%s\n", linuxStr)
			return nil
		}
	} else {
		t, err := credentials.PromptAndWrite(c.In(), c.Out(), api.PrefixURI.Host)
		if err != nil {
			if runtime.GOOS == "windows" {
				fmt.Printf("Unable to write credential.\n\nPlease execute the following, add it to your Powershell profile, or add it to your environment variables in control panel: \nWith Powershell:\n$env:SECTION_TOKEN=\"%s\"\n\nWith CMD:\nset SECTION_TOKEN=%s\n\nWith control panel:\n%s", t, t, screenshot)
				return nil
			}
			fmt.Printf("%s%s\n", linuxStr, t)
			return nil
		}
		api.Token = t
	}
	log.Info().Msg("Validating credentials...")
	_, err = api.CurrentUser()
	if err != nil {
		fmt.Println("error!")
		if errors.Is(err, api.ErrAuthDenied) {
			return err
		}
		return fmt.Errorf("could not fetch current user: %w", err)
	}
	log.Info().Msg(fmt.Sprintln("success!"))
	
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
