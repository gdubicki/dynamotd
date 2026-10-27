package plugins

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/gdubicki/dynamotd/dynamotd"
)

type commandSettings struct {
	Label string `yaml:"label"`
	// the command to run with /bin/sh
	Run string `yaml:"run"`
}

func (settings commandSettings) Validate() error {
	if settings.Label == "" {
		return errors.New("label is required")
	}
	if settings.Run == "" {
		return errors.New("run is required")
	}
	return nil
}

func (settings commandSettings) RowLabel() string {
	return settings.Label
}

// CommandRowType is the name of the row type running commands, which requires a secure config file.
const CommandRowType = "command"

func init() {
	dynamotd.Register("Command", commandSettings{}, Command, CommandRowType)
}

// how long to wait for the command's output after killing it on timeout,
// in case it has started subprocesses that keep its output open
const commandWaitDelay = 50 * time.Millisecond

// Command shows the first line of the output of a custom command. Its exit code determines the color,
// like for the Nagios plugins: 0 - normal, 1 - warning, 2 and other - critical.
func Command(ctx context.Context, settings commandSettings) []dynamotd.Row {
	cmd := exec.CommandContext(ctx, "/bin/sh", "-c", settings.Run)
	cmd.WaitDelay = commandWaitDelay
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	err := cmd.Run()

	if ctx.Err() != nil {
		return []dynamotd.Row{dynamotd.TimedOutRow(settings.Label)}
	}

	output := firstLine(stdout.String())

	valueColor := dynamotd.ValueNeutralColor
	if err != nil {
		var exitErr *exec.ExitError
		if !errors.As(err, &exitErr) {
			return []dynamotd.Row{dynamotd.ErrorRow(settings.Label, err)}
		}
		valueColor = exitCodeColor(exitErr.ExitCode())
		if output == "" {
			output = fmt.Sprintf("exit code %d", exitErr.ExitCode())
		}
	}
	if output == "" {
		output = "(no output)"
	}

	return []dynamotd.Row{{
		Label: dynamotd.SingleColorLabel(settings.Label),
		Value: dynamotd.ToColorText(dynamotd.ColorString{Color: valueColor, Text: output}),
	}}
}

func exitCodeColor(exitCode int) color.Attribute {
	if exitCode == 1 {
		return dynamotd.ValueWarningColor
	}
	return dynamotd.ValueCriticalColor
}

func firstLine(output string) string {
	for _, line := range strings.Split(output, "\n") {
		if line = strings.TrimSpace(line); line != "" {
			return line
		}
	}
	return ""
}
