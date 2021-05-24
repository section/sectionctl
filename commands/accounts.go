package commands

import (
	"os"
	"strconv"

	"github.com/alecthomas/kong"
	"github.com/section/sectionctl/api"
)

// AccountsCmd manages accounts on Section
type AccountsCmd struct {
	List AccountsListCmd `cmd help:"List accounts on Section." default:"1"`
}

// AccountsListCmd handles listing accounts on Section
type AccountsListCmd struct{}

// Run executes the command
func (c *AccountsListCmd) Run(cli *CLI, ctx *kong.Context,logWriters *LogWriters) (err error) {
	s := NewSpinner(cli, "Looking up accounts", logWriters)
	s.Start()

	accounts, err := api.Accounts()
	s.Stop()
	if err != nil {
		return err
	}

	table := NewTable(cli, os.Stdout)
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
