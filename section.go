package main

import (
	"fmt"
	"github.com/alecthomas/kong"
	"github.com/jdxcode/netrc"
	"os/user"
	"path/filepath"
	"runtime"
)

// CLI is the type wrapping all commands
var CLI struct {
	Login   LoginCmd  `cmd help:"Authenticate to Section's API."`
	Apps    AppsCmd   `cmd help:"Manage apps on Section"`
	Deploy  DeployCmd `cmd help:"Deploy an app to Section"`
	Version struct{}  `cmd help:"Print section-cli version"`
}

// AppsCmd manages apps on Section
type AppsCmd struct {
	List   AppsListCmd   `cmd help:"List apps on Section." default:"1"`
	Create AppsCreateCmd `cmd help:"Create new app on Section."`
}

// AppsListCmd handles listing apps running on Section
type AppsListCmd struct{}

// Run executes the `apps list` command
func (a *AppsListCmd) Run() (err error) {
	fmt.Println("omgwtfbbq")
	return err
}

// AppsCreateCmd handles creating apps on Section
type AppsCreateCmd struct{}

// LoginCmd handles authenticating the CLI against Section's API
type LoginCmd struct{}

// Run executes the `login` command
func (a *LoginCmd) Run() (err error) {
	usr, err := user.Current()
	n, err := netrc.Parse(filepath.Join(usr.HomeDir, ".netrc"))
	fmt.Println(n.Machine("aperture.section.io").Get("login"))
	fmt.Println(n.Machine("aperture.section.io").Get("password"))
	return err
}

// DeployCmd handles deploying an app to Section
type DeployCmd struct{}

func main() {
	ctx := kong.Parse(&CLI,
		kong.Description("CLI to interact with Section."),
		kong.UsageOnError(),
		kong.ConfigureHelp(kong.HelpOptions{Tree: true}),
	)
	switch ctx.Command() {
	case "login":
		ctx.Run()
	case "apps":
		fmt.Println("apps")
	case "apps create":
		fmt.Println("create an app")
	case "apps list":
		fmt.Println("list apps")
		ctx.Run()
	case "deploy":
		fmt.Println("deploy")
	case "version":
		fmt.Printf("%s (%s-%s)\n", "0.0.1", runtime.GOOS, runtime.GOARCH)
	}
}
