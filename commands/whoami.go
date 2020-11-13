package commands

import (
	"fmt"
	"os"
	"strconv"

	"github.com/section/sectionctl/api"
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
	u, err := api.CurrentUser()
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
