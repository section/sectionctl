package commands

import (
	"fmt"
	"github.com/jdxcode/netrc"
	"os/user"
	"path/filepath"
)

// LoginCmd handles authenticating the CLI against Section's API
type LoginCmd struct{}

// Run executes the `login` command
func (a *LoginCmd) Run() (err error) {
	usr, err := user.Current()
	if err != nil {
		panic(err)
	}
	n, err := netrc.Parse(filepath.Join(usr.HomeDir, ".netrc"))
	if err != nil {
		panic(err)
	}
	fmt.Println(n.Machine("aperture.section.io").Get("login"))
	fmt.Println(n.Machine("aperture.section.io").Get("password"))
	return err
}
