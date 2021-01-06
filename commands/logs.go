package commands

import (
	"fmt"
	"log"
	"strings"

	"github.com/mattn/go-colorable" // colorable
	"github.com/logrusorgru/aurora"
	"github.com/section/sectionctl/api"
)

// maxNumberLogs
const maxNumberLogs = 1500

// LogsCmd returns logs from an application on Section's delivery platform
type LogsCmd struct {
	AccountID       int    `required short:"a" help:"ID of account to query"`
	AppID           int    `required short:"i" help:"ID of app to query"`
	AppPath         string `default:"nodejs" help:"Path of NodeJS application in environment repository."`
	InstanceName    string `default:"" help:"Specific instance of NodeJS application running on Section platform."`
	Length          int    `default:100 help:"Number of log lines to fetch."`
	// StartTimestamp  int    `default:0 help:"Start of log time stamp to fetch."`
	// EndTimestamp    int    `default:0 help:"End of log time stamp to fetch."`
}

// Run executes the command
func (c *LogsCmd) Run() (err error) {
	s := NewSpinner("Getting logs from app")
	s.Start()

	if c.Length > maxNumberLogs {
		return fmt.Errorf("number of logs queried cannot be over %d", maxNumberLogs)
	}

	appLogs, err := api.ApplicationLogs(c.AccountID, c.AppID, c.AppPath, c.InstanceName, c.Length)
	s.Stop()
	if err != nil {
		return err
	}

	// Fix colorization issues between aurora and Windows
	// https://github.com/logrusorgru/aurora#windows
	log.SetOutput(colorable.NewColorableStdout())
	log.SetFlags(log.Flags() &^ (log.Ldate | log.Ltime)) // Remove local time prefix on output
	log.Printf("App InstanceName[Log Type]\t\tLog Message\n") 
	for _, a := range appLogs {
		a.Message = strings.TrimSpace(a.Message)
		
		if a.Type == "app" {
			log.Printf("%s%s\t%s\n", aurora.Cyan(a.InstanceName), aurora.Cyan("[" + a.Type + "]"), a.Message)
		} else if a.Type == "access" {
			log.Printf("%s%s\t%s\n", aurora.Green(a.InstanceName), aurora.Green("[" + a.Type + "]"), a.Message)
		} else {
			log.Printf("%s[%s]\t%s\n", a.InstanceName, a.Type, a.Message)
		}
	}
	
	return nil
}
