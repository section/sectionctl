package commands

import (
	"encoding/json"
	"fmt"

	"github.com/section/sectionctl/analytics"
)

// AnalyticsCmd handles recording analytics events
type AnalyticsCmd struct {
	Event string `required`
}

// Run executes the command
func (c *AnalyticsCmd) Run() (err error) {
	var e analytics.Event
	err = json.Unmarshal([]byte(c.Event), &e)
	if err != nil {
		return fmt.Errorf("unable to parse JSON event: %w", err)
	}
	err = analytics.Submit(e)
	return err
}
