package commands

import (
	"io"
	"os"
	"strconv"

	"github.com/olekukonko/tablewriter"
	"github.com/section/section-cli/api"
)

// AppsCmd manages apps on Section
type AppsCmd struct {
	List   AppsListCmd   `cmd help:"List apps on Section." default:"1"`
	Info   AppsInfoCmd   `cmd help:"Show detailed app information on Section."`
	Create AppsCreateCmd `cmd help:"Create new app on Section."`
}

// AppsListCmd handles listing apps running on Section
type AppsListCmd struct {
	APIBase   string `default:"https://aperture.section.io/api/v1"`
	AccountID int    `required short:"a"`
}

// NewTable returns a table with section-cli standard formatting
func NewTable(out io.Writer) (t *tablewriter.Table) {
	t = tablewriter.NewWriter(out)
	t.SetBorders(tablewriter.Border{Left: true, Top: false, Right: true, Bottom: false})
	t.SetCenterSeparator("|")
	t.SetAlignment(tablewriter.ALIGN_LEFT)
	return t
}

// Run executes the command
func (c *AppsListCmd) Run() (err error) {
	apps, err := api.Applications(c.AccountID)
	if err != nil {
		return err
	}

	table := NewTable(os.Stdout)
	table.SetHeader([]string{"App ID", "App Name"})

	for _, a := range apps {
		r := []string{strconv.Itoa(a.ID), a.ApplicationName}
		table.Append(r)
	}

	table.Render()
	return err
}

// AppsInfoCmd shows detailed information on an app running on Section
type AppsInfoCmd struct {
	AccountID     int `required short:"a"`
	ApplicationID int `required short:"i"`
}

// Run executes the command
func (c *AppsInfoCmd) Run() (err error) {
	app, err := api.Application(c.AccountID, c.ApplicationID)
	if err != nil {
		return err
	}

	table := NewTable(os.Stdout)
	table.SetHeader([]string{"App ID", "App Name", "Environments", "Domains"})
	table.SetAutoMergeCells(true)

	for _, env := range app.Environments {
		for _, dom := range env.Domains {
			r := []string{strconv.Itoa(app.ID), app.ApplicationName, env.EnvironmentName, dom.Name}
			table.Append(r)
		}
	}

	table.Render()
	return err
}

// AppsCreateCmd handles creating apps on Section
type AppsCreateCmd struct{}
