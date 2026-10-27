package plugins

import (
	"context"
	"testing"
	"time"

	"github.com/fatih/color"
	"github.com/gdubicki/dynamotd/dynamotd"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCommand(t *testing.T) {
	tests := []struct {
		name     string
		run      string
		expected string
		color    color.Attribute
	}{
		{"ok", "echo '  5 containers  '", "5 containers", dynamotd.ValueNeutralColor},
		{"first non-empty line only", "echo; echo first; echo second", "first", dynamotd.ValueNeutralColor},
		{"no output", "true", "(no output)", dynamotd.ValueNeutralColor},
		{"warning", "echo 'WARNING - 3 jobs queued'; exit 1", "WARNING - 3 jobs queued", dynamotd.ValueWarningColor},
		{"critical", "echo 'CRITICAL - down'; exit 2", "CRITICAL - down", dynamotd.ValueCriticalColor},
		{"other error without output", "echo oops >&2; exit 127", "exit code 127", dynamotd.ValueCriticalColor},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rows := Command(context.Background(), commandSettings{Label: "Custom", Run: tt.run})

			require.Len(t, rows, 1)
			assert.Equal(t, "Custom", rows[0].Label.String())
			assert.Equal(t, tt.expected, rows[0].Value.String())
			assert.Equal(t, tt.color, rows[0].Value.Parts()[0].Color)
		})
	}
}

func TestCommandTimeout(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	start := time.Now()
	// the subprocess keeps the output open even after the shell is killed
	rows := Command(ctx, commandSettings{Label: "Slow", Run: "sleep 10 & sleep 10"})

	assert.Less(t, time.Since(start), 2*time.Second)
	require.Len(t, rows, 1)
	assert.Equal(t, "timed out", rows[0].Value.String())
}

func TestCommandSettings(t *testing.T) {
	assert.ErrorContains(t, commandSettings{Run: "true"}.Validate(), "label is required")
	assert.ErrorContains(t, commandSettings{Label: "X"}.Validate(), "run is required")

	rowTypes, err := dynamotd.Lookup([]dynamotd.RowConfig{{Name: CommandRowType, Settings: settingsNodeFor(t, "label: Jobs\nrun: echo 1\n")}})
	require.NoError(t, err)
	// the label from the settings is used, f.e. for the timeout row
	assert.Equal(t, "Jobs", rowTypes[0].Label)
}
