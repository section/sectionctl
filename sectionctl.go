package main

import (
	"fmt"
	"io"
	golog "log"
	"os"
	"path/filepath"
	"time"

	"github.com/alecthomas/kong"
	"github.com/mattn/go-colorable"
	"github.com/posener/complete"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/section/sectionctl/api"
	"github.com/section/sectionctl/commands"
	"github.com/section/sectionctl/credentials"
	"github.com/willabides/kongplete"
)

//go:generate go-winres make --product-version=git-tag --file-version=git-tag
func bootstrap(c *commands.CLI, cmd *kong.Context) {

	api.PrefixURI = c.SectionAPIPrefix
	api.Timeout = c.SectionAPITimeout
	ctx := cmd
	colorableWriter := colorable.NewColorableStdout()
	consoleWriter := zerolog.ConsoleWriter{Out: colorableWriter, PartsExclude: []string{zerolog.TimestampFieldName,zerolog.LevelFieldName}}
	fileOutput := io.Discard
	multi := zerolog.MultiLevelWriter(consoleWriter)
	if len(c.DebugFile)>0 || bool(c.DebugOutput){
		if len(c.DebugFile) == 0 {
			c.DebugFile = commands.DebugFileFlag(fmt.Sprintf("sectionctl-debug-%s.log", time.Now().Format("2006-01-02-15-04-05Z0700")))
		}
		logFilePath := filepath.Join(string(c.DebugFile))
		logFile, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0666)
		if err != nil {
			panic(err)
		}
		fmt.Fprintf(logFile, "Version:   %s\n", commands.VersionCmd{}.String())
		fmt.Fprintf(logFile, "Command:   %s\n", ctx.Args)
		fmt.Fprintf(logFile, "PrefixURI: %s\n", api.PrefixURI)
		fmt.Fprintf(logFile, "Timeout:   %s\n", api.Timeout)
		fmt.Printf("Writing debug log to: %s\n", logFilePath)
		fileOutput = zerolog.New(logFile).With().Timestamp().Logger()
		if c.Quiet{
			zerolog.SetGlobalLevel(zerolog.ErrorLevel)
		}
		multi = zerolog.MultiLevelWriter(consoleWriter, fileOutput)
	}
	logger := zerolog.New(multi)
	golog.SetOutput(multi)
	log.Logger = logger
	logWriters := commands.LogWriters{
		ConsoleWriter: consoleWriter,
		FileWriter: fileOutput,
		ConsoleOnly: colorableWriter,
		CarriageReturnWriter: colorableWriter,
	}
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if c.Debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
		logWriters.CarriageReturnWriter = io.MultiWriter(logWriters.ConsoleOnly,logWriters.FileWriter);
	}
	if c.Quiet {
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
		logWriters.CarriageReturnWriter = logWriters.FileWriter
	}

	ctx.Bind(&logWriters)
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
				log.Fatal().Err(err)
			}
			t = to

		}
		api.Token = t
	}
}

func main() {
	// Handle completion requests
	var c commands.CLI
	parser := kong.Must(&c, kong.Name("sectionctl"), kong.UsageOnError())
	kongplete.Complete(parser,
		kongplete.WithPredictor("file", complete.PredictFiles("*")),
	)


	golog.SetFlags(0)
	cmd := kong.Parse(&c,
		kong.Description("CLI to interact with Section."),
		kong.UsageOnError(),
		kong.Bind(&c),
		kong.Configuration(commands.PackageJSONResolver, "package.json"),
		kong.ConfigureHelp(kong.HelpOptions{Tree: true}),
	)

	bootstrap(&c, cmd)

	
	er := cmd.Run()
	cmd.FatalIfErrorf(er)
}
