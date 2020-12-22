package commands

import (
	"io"
	"os"

	"github.com/olekukonko/tablewriter"
	"github.com/section/sectionctl/api"
)

// LogsCmd returns logs from an application on Section's delivery platform
type LogsCmd struct {
	AccountID       int    `required short:"a" help:"ID of account to query"`
	AppID           int    `required short:"i" help:"ID of app to query"`
	AppPath         string `default:"nodejs" help:"Path of NodeJS application in environment repository."`
	instanceName    string `default:"" help:"Specific instance of NodeJS application running on Section platform."`
	Length          int    `default:100 help:"Number of log lines to fetch."`
	// StartTimestamp  int    `default:0 help:"Start of log time stamp to fetch."`
	// EndTimestamp    int    `default:0 help:"End of log time stamp to fetch."`
}

// NewTable returns a table with sectionctl standard formatting
func NewLogTable(out io.Writer) (t *tablewriter.Table) {
	t = tablewriter.NewWriter(out)
	t.SetBorders(tablewriter.Border{Left: false, Top: false, Right: false, Bottom: false})
	t.SetColumnSeparator(" ")
	t.SetAlignment(tablewriter.ALIGN_LEFT)
	t.SetAutoWrapText(false)
	t.SetHeader([]string{"Timestamp", "App Instance", "Message"})
	t.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	t.SetHeaderLine(false)
	return t
}

// Run executes the command
func (c *LogsCmd) Run() (err error) {
	s := NewSpinner("Getting logs from app")
	s.Start()

	appLogs, err := api.ApplicationLogs(c.AccountID, c.AppID, c.AppPath, c.instanceName, c.Length)
	s.Stop()
	if err != nil {
		return err
	}

	table := NewLogTable(os.Stdout)
	for _, a := range appLogs {
		r := []string{a.Timestamp, a.Source + "[" + a.Type + "]", a.Message}
		if a.Type == "app" {
			table.Rich(r, []tablewriter.Colors{tablewriter.Colors{tablewriter.Normal, tablewriter.FgCyanColor}, tablewriter.Colors{tablewriter.Normal, tablewriter.FgCyanColor}, tablewriter.Colors{tablewriter.Normal, tablewriter.FgWhiteColor}})
		} else if a.Type == "access" {
			table.Rich(r, []tablewriter.Colors{tablewriter.Colors{tablewriter.Normal, tablewriter.FgGreenColor}, tablewriter.Colors{tablewriter.Normal, tablewriter.FgGreenColor}, tablewriter.Colors{tablewriter.Normal, tablewriter.FgWhiteColor}})
		} else {
			table.Append(r)
		}
	}
	table.Render()
	return nil
}
