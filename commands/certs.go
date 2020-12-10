package commands

import (
	"fmt"
	"time"

	"github.com/section/sectionctl/api"
	"github.com/section/sectionctl/api/auth"
)

// CertsCmd manages certificates on Section
type CertsCmd struct {
	Renew CertsRenewCmd `cmd help:"Renew a certificate for a domain."`
}

// CertsRenewCmd handles renewing a certificate
type CertsRenewCmd struct {
	AccountID int    `required short:"a"`
	Hostname  string `required`
}

// Run executes the command
func (c *CertsRenewCmd) Run() (err error) {
	s := NewSpinner()

	err = auth.Setup(api.PrefixURI.Host)
	if err != nil {
		return err
	}

	s.Suffix = fmt.Sprintf(" Renewing cert for %s", c.Hostname)
	s.Start()
	time.Sleep(2 * time.Second)

	resp, err := api.DomainsRenewCert(c.AccountID, c.Hostname)
	s.Stop()
	if err != nil {
		return err
	}

	fmt.Printf("\nSuccess: %s\n", resp.Message)

	return err
}
