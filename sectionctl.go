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
	DebugFile       	 string                       `default:"." help:"Directory where debug output should be written"`
	SectionToken       string                       `env:"SECTION_TOKEN" help:"Secret token for API auth"`
	SectionAPIPrefix   *url.URL                     `default:"https://aperture.section.io" env:"SECTION_API_PREFIX"`
	SectionAPITimeout  time.Duration                `default:"30s" env:"SECTION_API_TIMEOUT" help:"Request timeout for the Section API"`
	InstallCompletions kongplete.InstallCompletions `cmd:"" help:"install shell completions"`
	Quiet					     bool                         `env:"SECTION_CI" help:"Enables minimal logging, for use in continuous integration."`
}




func bootstrap(c CLI, cmd *kong.Context) context.Context {
	api.PrefixURI = c.SectionAPIPrefix
	api.Timeout = c.SectionAPITimeout

	ctx := context.Background()
	minLogLevel := logutils.LogLevel("INFO")
	if c.Quiet {
		ctx = context.WithValue(ctx, commands.CtxKey("quiet"), c.Quiet)
		minLogLevel = logutils.LogLevel("ERROR")
	}

	colorableWriter := colorable.NewColorableStderr()

	filter := &logutils.LevelFilter{
		Levels:   []logutils.LogLevel{"DEBUG", "INFO", "WARN", "ERROR"},
		MinLevel: minLogLevel,
		Writer:   colorableWriter,
	}
	if c.Debug {
		filter.MinLevel = logutils.LogLevel("DEBUG")
		if(c.DebugOutput || len(c.DebugFile) > 1){
			info, err := os.Stat(c.DebugFile); 
			if err != nil {
				panic(err)
			}
			logFilePath := filepath.Join(c.DebugFile)
			if len(c.DebugFile) == 0 || info.IsDir() {
				logFilePath = filepath.Join(c.DebugFile, fmt.Sprintf("sectionctl-debug-%s.log", time.Now().Format("2006-01-02-15-04-05Z0700")))
				c.DebugFile = logFilePath 
			}
			logFile, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0666)
			if err != nil {
				panic(err)
			}
			fmt.Fprintf(logFile, "Version:   %s\n", commands.VersionCmd{}.String())
			fmt.Fprintf(logFile, "Command:   %s\n", cmd.Args)
			fmt.Fprintf(logFile, "PrefixURI: %s\n", api.PrefixURI)
			fmt.Fprintf(logFile, "Timeout:   %s\n", api.Timeout)
			fmt.Printf("Writing debug log to: %s\n", logFilePath)
			mw := io.MultiWriter(logFile, colorableWriter)
			filter.Writer = mw
		}
	}
	log.SetOutput(filter)

	switch {
	case cmd.Command() == "version":
		// bypass auth check for version command
	case cmd.Command() == "login":
		api.Token = c.SectionToken
	case cmd.Command() != "login" && cmd.Command() != "logout":
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
	return ctx
}

func main() {
	// Handle completion requests
	var cli CLI
	parser := kong.Must(&cli, kong.Name("sectionctl"), kong.UsageOnError())
	kongplete.Complete(parser,
		kongplete.WithPredictor("file", complete.PredictFiles("*")),
	)

	cmd := kong.Parse(&cli,
		kong.Description("CLI to interact with Section."),
		kong.UsageOnError(),
		kong.ConfigureHelp(kong.HelpOptions{Tree: true}),
	)
	ctx := bootstrap(cli, cmd)
	cmd.BindTo(ctx, (*context.Context)(nil))
	err := cmd.Run(ctx)
	if err != nil {
		log.Printf("[ERROR] %s\n", err)
		os.Exit(2)
	}
}
