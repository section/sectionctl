package commands

import (
	"fmt"
	"time"

	"github.com/briandowns/spinner"
)

/*
// NewSpinner returns a nicely formatted spinner for display while users are waiting.
func NewSpinner() *spinner.Spinner {
	return spinner.New(spinner.CharSets[14], 100*time.Millisecond, spinner.WithWriter(os.Stderr))
}
*/

// NewSpinner returns a nicely formatted spinner for display while users are waiting.
func NewSpinner(cli *CLI, txt string, logWriters *LogWriters) (s *spinner.Spinner) {
	s = spinner.New(spinner.CharSets[35], 500*time.Millisecond, spinner.WithWriter(logWriters.CarriageReturnWriter))
	s.Color("cyan")
	s.Prefix = fmt.Sprintf("%s... ", txt)
	s.FinalMSG = fmt.Sprintf("%s... ✔️\n", txt)
	return s
}