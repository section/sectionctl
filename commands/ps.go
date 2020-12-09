package commands

import (
	"github.com/section/sectionctl/api"
	"github.com/section/sectionctl/api/auth"
	"os"
)

// PsCmd checks an application's status on Section's delivery platform
type PsCmd struct {
	AccountID int `required short:"a" help:"ID of account to query"`
	AppID     int `required short:"i" help:"ID of app to query"`
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
	err = auth.Setup(api.PrefixURI.Host)
	if err != nil {
		return err
	}
	appStatus, err := api.ApplicationStatus(c.AccountID, c.AppID)
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
