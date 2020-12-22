package commands

import (
	"os"

	"github.com/section/sectionctl/api"
)

// PsCmd checks an application's status on Section's delivery platform
type PsCmd struct {
	AccountID int    `required short:"a" help:"ID of account to query"`
	AppID     int    `required short:"i" help:"ID of app to query"`
	AppPath   string `default:"nodejs" help:"Path of NodeJS application in environment repository."`
}

func getStatus(as api.AppStatus) string {
	if as.InService && as.State == "Running" {
		return "Running"
	} else if as.State == "Deploying" {
		return "Deploying"
	} else {
		return "Not Running"
	}
}

// Run executes the command
func (c *PsCmd) Run() (err error) {
	s := NewSpinner("Getting status of app")
	s.Start()

	appStatus, err := api.ApplicationStatus(c.AccountID, c.AppID, c.AppPath)
	s.Stop()
	if err != nil {
		return err
	}

	table := NewTable(os.Stdout)
	table.SetHeader([]string{"App instance name", "App Status", "App Payload ID"})

	for _, a := range appStatus {
		r := []string{a.InstanceName, getStatus(a), a.PayloadID}
		table.Append(r)
	}
	table.Render()
	return nil
}
