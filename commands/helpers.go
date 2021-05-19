package commands

import (
	"context"
	"fmt"
	"io"
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

type CtxKey string

func IsInCtxBool(ctx context.Context, arg string) bool {
	return ctx.Value(CtxKey(arg)) != nil && ctx.Value(CtxKey(arg)).(bool)
}

// NewSpinner returns a nicely formatted spinner for display while users are waiting.
func NewSpinner(ctx context.Context, txt string) (s *spinner.Spinner) {
	var writer io.Writer
	if IsInCtxBool(ctx, "quiet"){
		writer = io.Discard
	}else{
		writer = os.Stderr
	}
	s = spinner.New(spinner.CharSets[35], 500*time.Millisecond, spinner.WithWriter(writer))
	s.Color("cyan")
	s.Prefix = fmt.Sprintf("%s... ", txt)
	s.FinalMSG = fmt.Sprintf("%s... ✔️\n", txt)
	return s
}