package commands

import (
	"fmt"
	"os"
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
func NewSpinner(txt string) (s *spinner.Spinner) {
	s = spinner.New(spinner.CharSets[14], 100*time.Millisecond, spinner.WithWriter(os.Stderr))
	s.Prefix = fmt.Sprintf("%s... ", txt)
	s.FinalMSG = fmt.Sprintf("%s... done\n", txt)
	return s
}
