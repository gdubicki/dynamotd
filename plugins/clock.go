package plugins

import (
	"context"

	"github.com/gdubicki/dynamotd/dynamotd"
)

func init() {
	dynamotd.Register(clockLabel, dynamotd.NoSettings{}, Clock, "clock")
}

const clockLabel = "Clock"

// for mocking in tests; returns ok=false if it can't be determined
var clockSynchronized = isClockSynchronized

// Clock shows that the clock is not synchronized, if it isn't, as that breaks f.e. TLS, Kerberos,
// log correlation, and other things in surprising ways.
func Clock(context.Context, dynamotd.NoSettings) []dynamotd.Row {
	synchronized, ok := clockSynchronized()
	if !ok || synchronized {
		return nil
	}
	return []dynamotd.Row{{
		Label: dynamotd.SingleColorLabel(clockLabel),
		Value: dynamotd.ToColorText(
			dynamotd.ValueCritical("not synchronized"),
			dynamotd.ValueDescription(" - check NTP"),
		),
	}}
}
