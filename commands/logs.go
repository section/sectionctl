package commands

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/logrusorgru/aurora"
	"github.com/mattn/go-colorable" // colorable
	"github.com/section/sectionctl/api"
)

// maxNumberLogs
const maxNumberLogs = 1500

// LogsCmd returns logs from an application on Section's delivery platform
type LogsCmd struct {
	AccountID    int    `required short:"a" help:"ID of account to query"`
	AppID        int    `required short:"i" help:"ID of app to query"`
	AppPath      string `default:"nodejs" help:"Path of NodeJS application in environment repository."`
	InstanceName string `default:"" help:"Specific instance of NodeJS application running on Section platform."`
	Number       int    `default:100 help:"Number of log lines to fetch."`
	Follow       bool   `help:"Displays recent logs and leaves the session open for logs to stream in. --instance-name required."`
	// StartTimestamp  int    `default:0 help:"Start of log time stamp to fetch."`
	// EndTimestamp    int    `default:0 help:"End of log time stamp to fetch."`
}

// Run executes the command
func (c *LogsCmd) Run() (err error) {
	s := NewSpinner("Getting logs from app")
	logsHeader := "\nInstanceName[Log Type]\t\t\tLog Message\n"
	s.FinalMSG = "done" + logsHeader
	s.Start()

	if c.Number > maxNumberLogs {
		return fmt.Errorf("number of logs queried cannot be over %d", maxNumberLogs)
	}

	var startTimestampRfc3339 string
	if c.Follow {
		log.Println("[DEBUG] Following logs...")
		if c.InstanceName == "" {
			return fmt.Errorf("--instance-name is required when using --follow")
		}
		startTimestampRfc3339 = time.Now().Format(time.RFC3339)
	}

	// Fix colorization issues between aurora and Windows
	// https://github.com/logrusorgru/aurora#windows
	log.SetOutput(colorable.NewColorableStdout())
	log.SetFlags(log.Flags() &^ (log.Ldate | log.Ltime)) // Remove local time prefix on output

	for {
		appLogs, err := api.ApplicationLogs(c.AccountID, c.AppID, c.AppPath, c.InstanceName, c.Number, startTimestampRfc3339)
		s.Stop()
		if err != nil {
			return err
		}
		var latestTimestamp string
		for _, a := range appLogs {
			a.Message = strings.TrimSpace(a.Message)

			if a.Type == "app" {
				log.Printf("%s%s\t%s\n", aurora.Cyan(a.InstanceName), aurora.Cyan("["+a.Type+"]"), a.Message)
			} else if a.Type == "access" {
				log.Printf("%s%s\t%s\n", aurora.Green(a.InstanceName), aurora.Green("["+a.Type+"]"), a.Message)
			} else {
				log.Printf("%s[%s]\t%s\n", a.InstanceName, a.Type, a.Message)
			}
			if a.Timestamp != "" {
				latestTimestamp = a.Timestamp
			}
		}
		if !c.Follow {
			break
		}
		if latestTimestamp == "" {
			latestTimestamp = startTimestampRfc3339
		}
		t, err := time.Parse(time.RFC3339, latestTimestamp)
		if err == nil {
			t = t.Add(time.Second)
			startTimestampRfc3339 = t.Format(time.RFC3339)
		}
		s.Prefix = ""
		s.FinalMSG = ""
		s.Start()
		time.Sleep(5 * time.Second)
	}
	return nil
}
