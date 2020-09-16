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

// Run executes the `apps list` command
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

// AppsCreateCmd handles creating apps on Section
type AppsCreateCmd struct{}
