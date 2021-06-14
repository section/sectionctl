package commands

import (
	"io"
	"net/url"
	"time"

	"github.com/fatih/color"
	"github.com/rs/zerolog"
	"github.com/willabides/kongplete"
)

// I would put this file in the 'main' package, but if I did that, then I wouldn't be able to reference CLI in the parameters of commands
// as golang doesn't like circular dependencies. Ideally all of our commands should be in the 'main' package in my opinion.

type debugFlag bool
type debugOutputFlag bool
type quietFlag bool
type DebugFileFlag string

func (d debugFlag) BeforeApply() error {
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	return nil
}
func (q quietFlag) BeforeApply() error {
	zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	return nil
}

// CLI exposes all the subcommands available
type CLI struct {
	Login              LoginCmd                     `cmd help:"Authenticate to Section's API"`
	Logout             LogoutCmd                    `cmd help:"Revoke authentication tokens to Section's API"`
	Accounts           AccountsCmd                  `cmd help:"Manage accounts on Section"`
	Apps               AppsCmd                      `cmd help:"Manage apps on Section"`
	Domains            DomainsCmd                   `cmd help:"Manage domains on Section"`
	Certs              CertsCmd                     `cmd help:"Manage certificates on Section"`
	Deploy             DeployCmd                    `cmd help:"Deploy an app to Section"`
	Logs               LogsCmd                      `cmd help:"Show logs from running applications"`
	Ps                 PsCmd                        `cmd help:"Show status of running applications"`
	Version            VersionCmd                   `cmd help:"Print sectionctl version"`
	WhoAmI             WhoAmICmd                    `cmd name:"whoami" help:"Show information about the currently authenticated user"`
	Debug              debugFlag                    `env:"DEBUG" default:"false" help:"Enable debug output"`
	DebugOutput        debugOutputFlag              `short:"out" help:"Enable logging on the debug output."`
	DebugFile          DebugFileFlag                `help:"File path where debug output should be written"`
	SectionToken       string                       `env:"SECTION_TOKEN" help:"Secret token for API auth"`
	SectionAPIPrefix   *url.URL                     `default:"https://aperture.section.io" env:"SECTION_API_PREFIX"`
	SectionAPITimeout  time.Duration                `default:"30s" env:"SECTION_API_TIMEOUT" help:"Request timeout for the Section API"`
	InstallCompletions kongplete.InstallCompletions `cmd:"" help:"install shell completions"`
	Quiet              quietFlag                    `env:"SECTION_CI" help:"Enables minimal logging, for use in continuous integration."`
}

type LogWriters struct {
	ConsoleWriter        io.Writer
	FileWriter           io.Writer
	ConsoleOnly          io.Writer
	CarriageReturnWriter io.Writer
}

var (
	Green = color.New(color.Bold, color.FgGreen).SprintfFunc()
	HiGreen = color.New(color.Bold, color.FgHiGreen).SprintfFunc()
	Cyan = color.New(color.Bold, color.FgCyan).SprintfFunc()
	HiCyan = color.New(color.Bold, color.FgHiCyan).SprintfFunc()
	Magenta = color.New(color.Bold, color.FgMagenta).SprintfFunc()
	HiMagenta = color.New(color.Bold, color.FgHiMagenta).SprintfFunc()
	Red = color.New(color.Bold, color.FgRed).SprintfFunc()
	HiRed = color.New(color.Bold, color.FgHiRed).SprintfFunc()
	Blue = color.New(color.Bold, color.FgBlue).SprintfFunc()
	HiBlue = color.New(color.Bold, color.FgHiBlue).SprintfFunc()
	HiWhite = color.New(color.Bold, color.FgHiWhite).SprintfFunc()
	White = color.New(color.Bold, color.FgWhite).SprintfFunc()
	Yellow = color.New(color.Bold, color.FgYellow).SprintfFunc()
	HiYellow = color.New(color.Bold, color.FgHiYellow).SprintfFunc()
)
