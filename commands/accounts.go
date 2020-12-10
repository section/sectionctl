package commands

import (
	"os"
	"strconv"
	"time"

	"github.com/section/sectionctl/api"
	"github.com/section/sectionctl/api/auth"
)

// AccountsCmd manages accounts on Section
type AccountsCmd struct {
	List AccountsListCmd `cmd help:"List accounts on Section." default:"1"`
}

// AccountsListCmd handles listing accounts on Section
type AccountsListCmd struct{}

// Run executes the command
func (c *AccountsListCmd) Run() (err error) {
	s := NewSpinner()

	err = auth.Setup(api.PrefixURI.Host)
	if err != nil {
		return err
	}

	s.Suffix = " Looking up accounts..."
	s.Start()
	time.Sleep(1 * time.Second)

	accounts, err := api.Accounts()
	s.Stop()
	if err != nil {
		return err
	}

	table := NewTable(os.Stdout)
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
