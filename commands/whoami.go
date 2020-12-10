package commands

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/section/sectionctl/api"
	"github.com/section/sectionctl/api/auth"
)

// WhoAmICmd returns information about the currently authenticated user
type WhoAmICmd struct{}

// PrettyBool pretty prints a bool value
func PrettyBool(b bool) (s string) {
	if b {
		return "✔"
	}
	return "✘"
}

// Run executes the command
func (c *WhoAmICmd) Run() (err error) {
	s := NewSpinner()

	err = auth.Setup(api.PrefixURI.Host)
	if err != nil {
		return err
	}

	s.Suffix = " Looking up current user..."
	s.Start()
	time.Sleep(1 * time.Second)

	u, err := api.CurrentUser()
	s.Stop()
	if err != nil {
		return err
	}

	table := NewTable(os.Stdout)
	table.SetHeader([]string{"Attribute", "Value"})
	r := [][]string{
		[]string{"Name", fmt.Sprintf("%s %s", u.FirstName, u.LastName)},
		[]string{"Email", u.Email},
		[]string{"ID", strconv.Itoa(u.ID)},
		[]string{"Company", u.CompanyName},
		[]string{"Phone Number", u.PhoneNumber},
		[]string{"Verified?", PrettyBool(u.Verified)},
		[]string{"Requires 2FA?", PrettyBool(u.Requires2FA)},
	}
	table.AppendBulk(r)
	table.Render()

	return nil
}
