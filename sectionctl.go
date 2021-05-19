package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"github.com/alecthomas/kong"
	"github.com/hashicorp/logutils"
	"github.com/mattn/go-colorable"
	"github.com/posener/complete"
	"github.com/section/sectionctl/api"
	"github.com/section/sectionctl/commands"
	"github.com/section/sectionctl/credentials"
	"github.com/willabides/kongplete"
)

// CLI exposes all the subcommands available
type CLI struct {
	Login              commands.LoginCmd            `cmd help:"Authenticate to Section's API"`
	Logout             commands.LogoutCmd           `cmd help:"Revoke authentication tokens to Section's API"`
	Accounts           commands.AccountsCmd         `cmd help:"Manage accounts on Section"`
	Apps               commands.AppsCmd             `cmd help:"Manage apps on Section"`
	Domains            commands.DomainsCmd          `cmd help:"Manage domains on Section"`
	Certs              commands.CertsCmd            `cmd help:"Manage certificates on Section"`
	Deploy             commands.DeployCmd           `cmd help:"Deploy an app to Section"`
	Logs               commands.LogsCmd             `cmd help:"Show logs from running applications"`
	Ps                 commands.PsCmd               `cmd help:"Show status of running applications"`
	Version            commands.VersionCmd          `cmd help:"Print sectionctl version"`
	WhoAmI             commands.WhoAmICmd           `cmd name:"whoami" help:"Show information about the currently authenticated user"`
	Debug              bool                         `env:"DEBUG" help:"Enable debug output"`
	DebugOutput        bool                         `short:"out" help:"Enable logging on the debug output."`
	DebugFileDir       string                       `default:"." help:"Directory where debug output should be written"`
	SectionToken       string                       `env:"SECTION_TOKEN" help:"Secret token for API auth"`
	SectionAPIPrefix   *url.URL                     `default:"https://aperture.section.io" env:"SECTION_API_PREFIX"`
	SectionAPITimeout  time.Duration                `default:"30s" env:"SECTION_API_TIMEOUT" help:"Request timeout for the Section API"`
	InstallCompletions kongplete.InstallCompletions `cmd:"" help:"install shell completions"`
	Quiet					     bool                         `env:"SECTION_CI" help:"Enables minimal logging, for use in continuous integration."`
}




func bootstrap(c CLI, ctx *kong.Context) context.Context {
	api.PrefixURI = c.SectionAPIPrefix
	api.Timeout = c.SectionAPITimeout

	contxt := context.Background()
	contxt = context.WithValue(contxt, commands.CTXKEY("quiet"), c.Quiet)

	colorableWriter := colorable.NewColorableStderr()

	minLogLevel := logutils.LogLevel("INFO")
	if contxt.Value(commands.CTXKEY("quiet")).(bool) {
		minLogLevel = logutils.LogLevel("ERROR")
	}
	filter := &logutils.LevelFilter{
		Levels:   []logutils.LogLevel{"DEBUG", "INFO", "WARN", "ERROR"},
		MinLevel: minLogLevel,
		Writer:   colorableWriter,
	}
	if c.Debug {
		filter.MinLevel = logutils.LogLevel("DEBUG")
		if(c.DebugOutput){
			logFilename := fmt.Sprintf("sectionctl-debug-%s.log", time.Now().Format("2006-01-02-15-04-05Z0700"))
			logFilePath := filepath.Join(c.DebugFileDir, logFilename)
			logFile, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0666)
			if err != nil {
				panic(err)
			}
			fmt.Fprintf(logFile, "Version:   %s\n", commands.VersionCmd{}.String())
			fmt.Fprintf(logFile, "Command:   %s\n", ctx.Args)
			fmt.Fprintf(logFile, "PrefixURI: %s\n", api.PrefixURI)
			fmt.Fprintf(logFile, "Timeout:   %s\n", api.Timeout)
			fmt.Printf("Writing debug log to: %s\n", logFilePath)
			mw := io.MultiWriter(logFile, colorableWriter)
			filter.Writer = mw
		}
	}
	log.SetOutput(filter)

	switch {
	case ctx.Command() == "version":
		// bypass auth check for version command
	case ctx.Command() == "login":
		api.Token = c.SectionToken
	case ctx.Command() != "login" && ctx.Command() != "logout":
		t := c.SectionToken
		if t == "" {
			to, err := credentials.Setup(api.PrefixURI.Host)
			if err != nil {
				log.Fatalf("[ERROR] %s\n", err)
			}
			t = to

		}
		api.Token = t
	}
	return contxt
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
	contxt := bootstrap(cli, ctx)
	ctx.BindTo(contxt, (*context.Context)(nil))
	err := ctx.Run(contxt)
	if err != nil {
		log.Printf("[ERROR] %s\n", err)
		os.Exit(2)
	}
}
