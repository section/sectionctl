package commands

import (
	"fmt"
	"io"
	"os"
	"runtime"

	"github.com/section/sectionctl/version"
)

// VersionCmd handles authenticating the CLI against Section's API
type VersionCmd struct {
	in  io.Reader
	out io.Writer
}

// Run executes the `login` command
func (c *VersionCmd) Run() (err error) {
	latest := make(chan string, 1)
	go c.checkVersion(latest)

	fmt.Printf("%s (%s-%s)\n", version.Version, runtime.GOOS, runtime.GOARCH)

	v := <-latest
	fmt.Println(v)

	return err
}

func (c *VersionCmd) checkVersion(latest chan string) {
	latest <- "hello"
}

// In returns the input to read from
func (c *VersionCmd) In() io.Reader {
	if c.in != nil {
		return c.in
	}
	return os.Stdin
}

// Out returns the output to write to
func (c *VersionCmd) Out() io.Writer {
	if c.out != nil {
		return c.out
	}
	return os.Stdout
}
