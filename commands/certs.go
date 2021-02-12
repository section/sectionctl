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
	Hostname string `arg help:"The domain name to renew the cert for"`
}

// Run executes the command
func (c *CertsRenewCmd) Run() (err error) {
	var aid int
	s := NewSpinner("Looking up accounts")
	s.Start()

	as, err := api.Accounts()
	if err != nil {
		return fmt.Errorf("unable to look up accounts: %w", err)
	}

	for _, a := range as {
		ds, err := api.Domains(a.ID)
		if err != nil {
			return fmt.Errorf("unable to look up domains under account ID %d: %w", a.ID, err)
		}
		for _, d := range ds {
			if d.DomainName == c.Hostname {
				aid = a.ID
				s.Stop()
				break
			}
		}
	}
	s.Stop()

	if aid == 0 {
		return fmt.Errorf("unable to find the domain '%s' under any of your accounts.\n\nTry running `sectionctl domains` to see all your domains", c.Hostname)
	}

	s = NewSpinner(fmt.Sprintf("Renewing cert for %s", c.Hostname))
	s.Start()

	resp, err := api.DomainsRenewCert(aid, c.Hostname)
	s.Stop()
	if err != nil {
		return err
	}

	fmt.Printf("\nSuccess: %s\n", resp.Message)

	return err
}
