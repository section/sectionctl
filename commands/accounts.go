package commands

import (
	"context"
	"os"
	"strconv"

	"github.com/section/sectionctl/api"
)

// AccountsCmd manages accounts on Section
type AccountsCmd struct {
	List AccountsListCmd `cmd help:"List accounts on Section." default:"1"`
}

// AccountsListCmd handles listing accounts on Section
type AccountsListCmd struct{}

// Run executes the command
func (c *AccountsListCmd) Run(ctx context.Context) (err error) {
	s := NewSpinner(ctx, "Looking up accounts")
	s.Start()

	accounts, err := api.Accounts()
	s.Stop()
	if err != nil {
		return err
	}

	table := NewTable(ctx, os.Stdout)
	table.SetHeader([]string{"Account ID", "Account Name"})

	for _, a := range accounts {
		r := []string{strconv.Itoa(a.ID), a.AccountName}
		table.Append(r)
	}

	table.Render()
	return err
}

// AccountsCreateCmd handles creating apps on Section
type AccountsCreateCmd struct{}
