package commands

import (
	"fmt"
	"os"
	"time"

	"github.com/briandowns/spinner"
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
	err = auth.Setup(api.PrefixURI.Host)
	if err != nil {
		return err
	}

	s := spinner.New(spinner.CharSets[14], 100*time.Millisecond, spinner.WithWriter(os.Stderr))
	s.Suffix = fmt.Sprintf(" Renewing cert for %s", c.Hostname)
	s.Start()
	time.Sleep(2 * time.Second)

	resp, err := api.DomainsRenewCert(c.AccountID, c.Hostname)
	if err != nil {
		s.Stop()
		return err
	}

	fmt.Printf("\nSuccess: %s\n", resp.Message)

	return err
}
