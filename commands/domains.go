package commands

import (
	"fmt"
	"os"

	"github.com/section/sectionctl/api"
)

// DomainsCmd manages domains on Section
type DomainsCmd struct {
	List DomainsListCmd `cmd help:"List domains on Section." default:"1"`
}

// DomainsListCmd handles listing domains on Section
type DomainsListCmd struct {
	AccountID int `required short:"a"`
}

// Run executes the command
func (c *DomainsListCmd) Run() (err error) {
	s := NewSpinner("Looking up domains")
	s.Start()

	domains, err := api.Domains(c.AccountID)
	s.Stop()
	if err != nil {
		return err
	}

	table := NewTable(os.Stdout)
	table.SetHeader([]string{"Domain", "Engaged"})

	for _, d := range domains {
		r := []string{d.DomainName, fmt.Sprintf("%t", d.Engaged)}
		table.Append(r)
	}
	table.Render()

	return err
}
