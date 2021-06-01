package commands

import (
	"fmt"
	"time"

	"github.com/briandowns/spinner"
	"github.com/rs/zerolog/log"
)

/*
// NewSpinner returns a nicely formatted spinner for display while users are waiting.
func NewSpinner() *spinner.Spinner {
	return spinner.New(spinner.CharSets[14], 100*time.Millisecond, spinner.WithWriter(os.Stderr))
}
*/

// NewSpinner returns a nicely formatted spinner for display while users are waiting.
func NewSpinner(txt string, logWriters *LogWriters) (s *spinner.Spinner) {
	log.Debug().Msg(txt)
	s = spinner.New(spinner.CharSets[14], 450*time.Millisecond, spinner.WithWriter(logWriters.ConsoleOnly))
	err := s.Color("cyan")
	if err != nil {
		// have an internal fit about it
		log.Debug().Msg("couldn't set the color on the spinner 🥺")
	}
	s.Prefix = fmt.Sprintf("%s... ", txt)
	s.FinalMSG = fmt.Sprintf("%s... ✔️\n", txt)
	return s
}