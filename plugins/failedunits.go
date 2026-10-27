package plugins

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/gdubicki/dynamotd/dynamotd"
)

func init() {
	dynamotd.Register(failedUnitsLabel, dynamotd.NoSettings{}, FailedUnits, "failedunits")
}

const failedUnitsLabel = "Failed units"

// show at most this many unit names, to keep the row short
const maxFailedUnitsShown = 5

// for mocking in tests
var systemdRunning = func() bool {
	// the same check as sd_booted() does
	_, err := os.Stat("/run/systemd/system")
	return err == nil
}

var listFailedUnits = func(ctx context.Context) ([]string, error) {
	output, err := exec.CommandContext(ctx, findCommand("systemctl"),
		"list-units", "--failed", "--plain", "--no-legend", "--no-pager").Output()
	if err != nil {
		return nil, err
	}
	var units []string
	for _, line := range strings.Split(string(output), "\n") {
		if fields := strings.Fields(line); len(fields) > 0 {
			units = append(units, fields[0])
		}
	}
	return units, nil
}

// findCommand returns the path to a system command, also when PATH is not set,
// which is possible when running from PAM.
func findCommand(name string) string {
	if path, err := exec.LookPath(name); err == nil {
		return path
	}
	for _, dir := range []string{"/usr/bin", "/bin", "/usr/sbin", "/sbin"} {
		path := dir + "/" + name
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}
	return name
}

// FailedUnits shows the failed systemd units, if there are any.
func FailedUnits(ctx context.Context, _ dynamotd.NoSettings) []dynamotd.Row {
	if !systemdRunning() {
		return nil
	}

	units, err := listFailedUnits(ctx)
	if err != nil {
		return []dynamotd.Row{dynamotd.ErrorRow(failedUnitsLabel, err)}
	}
	if len(units) == 0 {
		return nil
	}

	shown := units[:min(len(units), maxFailedUnitsShown)]
	value := []dynamotd.ColorString{
		dynamotd.ValueCritical(fmt.Sprintf("%d", len(units))),
		dynamotd.ValueDescription(": "),
		dynamotd.ValueCritical(strings.Join(shown, ", ")),
	}
	if len(units) > len(shown) {
		value = append(value, dynamotd.ValueDescription(fmt.Sprintf(" and %d more", len(units)-len(shown))))
	}

	return []dynamotd.Row{{
		Label: dynamotd.SingleColorLabel(failedUnitsLabel),
		Value: dynamotd.ToColorText(value...),
	}}
}
