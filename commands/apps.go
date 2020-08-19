package commands

import (
	"fmt"
)

// AppsCmd manages apps on Section
type AppsCmd struct {
	List   AppsListCmd   `cmd help:"List apps on Section." default:"1"`
	Create AppsCreateCmd `cmd help:"Create new app on Section."`
}

// AppsListCmd handles listing apps running on Section
type AppsListCmd struct{}

// Run executes the `apps list` command
func (c *AppsListCmd) Run() (err error) {
	fmt.Println("omgwtfbbq")
	return err
}

// AppsCreateCmd handles creating apps on Section
type AppsCreateCmd struct{}
