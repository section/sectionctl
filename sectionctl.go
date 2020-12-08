package main

import (
	"log"
	"net/url"
	"os"

	"github.com/alecthomas/kong"
	"github.com/hashicorp/logutils"
	"github.com/posener/complete"
	"github.com/section/sectionctl/analytics"
	"github.com/section/sectionctl/api"
	"github.com/section/sectionctl/commands"
	"github.com/willabides/kongplete"
)

// CLI exposes all the subcommands available
type CLI struct {
	Login              commands.LoginCmd            `cmd help:"Authenticate to Section's API."`
	Accounts           commands.AccountsCmd         `cmd help:"Manage accounts on Section"`
	Apps               commands.AppsCmd             `cmd help:"Manage apps on Section"`
	Certs              commands.CertsCmd            `cmd help:"Manage certificates on Section"`
	Deploy             commands.DeployCmd           `cmd help:"Deploy an app to Section"`
	Version            commands.VersionCmd          `cmd help:"Print sectionctl version"`
	WhoAmI             commands.WhoAmICmd           `cmd name:"whoami" help:"Show information about the currently authenticated user"`
	Ps                 commands.PsCmd               `cmd help:"Show status of running applications"`
	Debug              bool                         `env:"DEBUG" help:"Enable debug output"`
	SectionToken       string                       `env:"SECTION_TOKEN" help:"Secret token for API auth"`
	SectionAPIPrefix   *url.URL                     `default:"https://aperture.section.io" env:"SECTION_API_PREFIX"`
	InstallCompletions kongplete.InstallCompletions `cmd:"" help:"install shell completions"`
}

func bootstrap(c CLI) {
	api.Debug = c.Debug
	api.PrefixURI = c.SectionAPIPrefix
	api.Token = c.SectionToken

	filter := &logutils.LevelFilter{
		Levels:   []logutils.LogLevel{"DEBUG", "INFO", "WARN", "ERROR"},
		MinLevel: logutils.LogLevel("INFO"),
		Writer:   os.Stderr,
	}
	if c.Debug {
		filter.MinLevel = logutils.LogLevel("DEBUG")
	}
	log.SetOutput(filter)
}

func main() {
	// Handle completion requests
	var cli CLI
	parser := kong.Must(&cli, kong.Name("sectionctl"), kong.UsageOnError())
	kongplete.Complete(parser,
		kongplete.WithPredictor("file", complete.PredictFiles("*")),
	)

	ctx := kong.Parse(&cli,
		kong.Description("CLI to interact with Section."),
		kong.UsageOnError(),
		kong.ConfigureHelp(kong.HelpOptions{Tree: true}),
	)
	bootstrap(cli)
	analytics.LogInvoke(ctx)
	err := ctx.Run()
	if err != nil {
		log.Printf("[ERROR] %s\n", err)
		os.Exit(2)
	}
}
