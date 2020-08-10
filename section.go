package main

import (
	"github.com/alecthomas/kong"
	"github.com/section/section-cli/commands"
)

// CLI exposes all the subcommands available
var CLI struct {
	Login   commands.LoginCmd   `cmd help:"Authenticate to Section's API."`
	Apps    commands.AppsCmd    `cmd help:"Manage apps on Section"`
	Deploy  commands.DeployCmd  `cmd help:"Deploy an app to Section"`
	Version commands.VersionCmd `cmd help:"Print section-cli version"`
}

func main() {
	ctx := kong.Parse(&CLI,
		kong.Description("CLI to interact with Section."),
		kong.UsageOnError(),
		kong.ConfigureHelp(kong.HelpOptions{Tree: true}),
	)
	err := ctx.Run()
	if err != nil {
		panic(err)
	}
}
