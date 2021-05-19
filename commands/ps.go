package commands

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/section/sectionctl/api"
)

// PsCmd checks an application's status on Section's delivery platform
type PsCmd struct {
	AccountID int           `short:"a" help:"ID of account to query"`
	AppID     int           `short:"i" help:"ID of app to query"`
	AppPath   string        `default:"nodejs" help:"Path of NodeJS application in environment repository."`
	Watch     bool          `short:"w" help:"Run repeatedly, output status"`
	Interval  time.Duration `short:"t" default:"10s" help:"Interval to poll if watching"`
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
func (c *PsCmd) Run(ctx context.Context) (err error) {
	var aids []int
	if c.AccountID == 0 {
		s := NewSpinner(ctx, "Looking up accounts")
		s.Start()

		as, err := api.Accounts()
		if err != nil {
			return fmt.Errorf("unable to look up accounts: %w", err)
		}
		for _, a := range as {
			aids = append(aids, a.ID)
		}

		s.Stop()
	} else {
		aids = append(aids, c.AccountID)
	}

	var targets [][]int
	for _, id := range aids {
		if c.AppID == 0 {
			s := NewSpinner(ctx, "Looking up applications")
			s.Start()

			as, err := api.Applications(id)
			if err != nil {
				return fmt.Errorf("unable to look up applications: %w", err)
			}
			for _, a := range as {
				targets = append(targets, []int{id, a.ID})
			}

			s.Stop()
		} else {
			targets = append(targets, []int{id, c.AppID})
		}
	}

	if c.Watch {
		ticker := time.NewTicker(c.Interval)
		for ; true; <-ticker.C {
			err = pollAndOutput(ctx, targets, c.AppPath)
			if err != nil {
				return err
			}
		}
	} else {
		err = pollAndOutput(ctx, targets, c.AppPath)
		return err
	}

	return nil
}

func pollAndOutput(ctx context.Context, targets [][]int, appPath string) error {
	s := NewSpinner(ctx, "Getting status of apps")
	s.Start()

	table := NewTable(ctx, os.Stdout)
	table.SetHeader([]string{"Account ID", "App ID", "App instance name", "App Status", "App Payload ID"})

	for _, t := range targets {
		appStatus, err := api.ApplicationStatus(t[0], t[1], appPath)
		s.Stop()
		if err != nil {
			return err
		}

		for _, a := range appStatus {
			r := []string{
				strconv.Itoa(t[0]),
				strconv.Itoa(t[1]),
				a.InstanceName,
				getStatus(a),
				a.PayloadID,
			}
			table.Append(r)
		}
	}

	table.Render()
	return nil
}
