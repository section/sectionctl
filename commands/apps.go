package commands

import (
	"fmt"
	"io"
	"os"
	"strconv"

	"github.com/olekukonko/tablewriter"
	"github.com/section/sectionctl/api"
	"github.com/section/sectionctl/api/auth"
)

// AppsCmd manages apps on Section
type AppsCmd struct {
	List   AppsListCmd   `cmd help:"List apps on Section." default:"1"`
	Info   AppsInfoCmd   `cmd help:"Show detailed app information on Section."`
	Create AppsCreateCmd `cmd help:"Create new app on Section."`
}

// AppsListCmd handles listing apps running on Section
type AppsListCmd struct {
	AccountID int `required short:"a"`
}

// NewTable returns a table with sectionctl standard formatting
func NewTable(out io.Writer) (t *tablewriter.Table) {
	t = tablewriter.NewWriter(out)
	t.SetBorders(tablewriter.Border{Left: true, Top: false, Right: true, Bottom: false})
	t.SetCenterSeparator("|")
	t.SetAlignment(tablewriter.ALIGN_LEFT)
	return t
}

// Run executes the command
func (c *AppsListCmd) Run() (err error) {
	s := NewSpinner()

	err = auth.Setup(api.PrefixURI.Host)
	if err != nil {
		return err
	}

	s.Suffix = " Looking up apps..."
	s.Start()

	apps, err := api.Applications(c.AccountID)
	if err != nil {
		s.Stop()
		return err
	}
	s.Stop()

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
	AccountID int `required short:"a"`
	AppID     int `required short:"i"`
}

// Run executes the command
func (c *AppsInfoCmd) Run() (err error) {
	s := NewSpinner()

	err = auth.Setup(api.PrefixURI.Host)
	if err != nil {
		return err
	}

	s.Suffix = " Looking up info about app..."
	s.Start()

	app, err := api.Application(c.AccountID, c.AppID)
	if err != nil {
		s.Stop()
		return err
	}
	s.Stop()

	fmt.Printf("🌎🌏🌍\n")
	fmt.Printf("App Name: %s\n", app.ApplicationName)
	fmt.Printf("App ID: %d\n", app.ID)
	fmt.Printf("Environment count: %d\n", len(app.Environments))

	for i, env := range app.Environments {
		fmt.Printf("\n-----------------\n\n")
		fmt.Printf("Environment #%d: %s (ID:%d)\n\n", i+1, env.EnvironmentName, env.ID)
		fmt.Printf("💬 Domains (%d total)\n", len(env.Domains))

		for _, dom := range env.Domains {
			fmt.Println()

			table := NewTable(os.Stdout)
			table.SetHeader([]string{"Attribute", "Value"})
			table.SetAutoMergeCells(true)
			r := [][]string{
				[]string{"Domain name", dom.Name},
				[]string{"Zone name", dom.ZoneName},
				[]string{"CNAME", dom.CNAME},
				[]string{"Mode", dom.Mode},
			}
			table.AppendBulk(r)
			table.Render()
		}

		fmt.Println()
		mod := "modules"
		if len(env.Stack) == 1 {
			mod = "module"
		}
		fmt.Printf("🥞 Stack (%d %s total)\n", len(env.Stack), mod)
		fmt.Println()

		table := NewTable(os.Stdout)
		table.SetHeader([]string{"Name", "Image"})
		table.SetAutoMergeCells(true)
		for _, p := range env.Stack {
			r := []string{p.Name, p.Image}
			table.Append(r)
		}
		table.Render()
	}

	fmt.Println()

	return err
}

// AppsCreateCmd handles creating apps on Section
type AppsCreateCmd struct{}
