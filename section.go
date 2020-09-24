package main

import (
	"fmt"
	"net/url"
	"os"

	"github.com/alecthomas/kong"
	"github.com/section/sectionctl/analytics"
	"github.com/section/sectionctl/api"
	"github.com/section/sectionctl/commands"
)

// CLI exposes all the subcommands available
var CLI struct {
	Login            commands.LoginCmd    `cmd help:"Authenticate to Section's API."`
	Accounts         commands.AccountsCmd `cmd help:"Manage accounts on Section"`
	Apps             commands.AppsCmd     `cmd help:"Manage apps on Section"`
	Deploy           commands.DeployCmd   `cmd help:"Deploy an app to Section"`
	Version          commands.VersionCmd  `cmd help:"Print sectionctl version"`
	SectionAPIPrefix *url.URL             `default:"https://aperture.section.io" env:"SECTION_API_PREFIX"`
}

func main() {
	ctx := kong.Parse(&CLI,
		kong.Description("CLI to interact with Section."),
		kong.UsageOnError(),
		kong.ConfigureHelp(kong.HelpOptions{Tree: true}),
	)
	api.PrefixURI = CLI.SectionAPIPrefix
	analytics.LogInvoke(ctx)
	err := ctx.Run()
	if err != nil {
		fmt.Printf("\nError: %s\n", err)
		os.Exit(2)
	}
}
