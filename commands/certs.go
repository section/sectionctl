package commands

import (
	"fmt"

	"github.com/section/sectionctl/api"
)

// CertsCmd manages certificates on Section
type CertsCmd struct {
	Renew CertsRenewCmd `cmd help:"Renew a certificate for a domain."`
}

// CertsRenewCmd handles renewing a certificate
type CertsRenewCmd struct {
	AccountID int    `required short:"a" help:"Account ID the domain belongs to"`
	Hostname  string `required`
}

// Run executes the command
func (c *CertsRenewCmd) Run() (err error) {
	s := NewSpinner(fmt.Sprintf("Renewing cert for %s", c.Hostname))
	s.Start()

	resp, err := api.DomainsRenewCert(c.AccountID, c.Hostname)
	s.Stop()
	if err != nil {
		return err
	}

	fmt.Printf("\nSuccess: %s\n", resp.Message)

	return err
}
