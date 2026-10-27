package plugins

import (
	"context"
	"errors"
	"testing"

	"github.com/gdubicki/dynamotd/dynamotd"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func mockSystemd(t *testing.T, running bool, units []string, err error) {
	origRunning, origList := systemdRunning, listFailedUnits
	t.Cleanup(func() { systemdRunning, listFailedUnits = origRunning, origList })

	systemdRunning = func() bool { return running }
	listFailedUnits = func(context.Context) ([]string, error) { return units, err }
}

func TestFailedUnits(t *testing.T) {
	tests := []struct {
		name     string
		running  bool
		units    []string
		err      error
		expected []string
	}{
		{"no systemd", false, []string{"a.service"}, nil, nil},
		{"no failed units", true, nil, nil, nil},
		{"some failed units", true, []string{"nginx.service", "backup.timer"}, nil,
			[]string{"2: nginx.service, backup.timer"}},
		{"many failed units", true, []string{"a", "b", "c", "d", "e", "f", "g"}, nil,
			[]string{"7: a, b, c, d, e and 2 more"}},
		{"error", true, nil, errors.New("boom"), []string{"can't get: boom"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSystemd(t, tt.running, tt.units, tt.err)

			rows := FailedUnits(context.Background(), dynamotd.NoSettings{})

			var values []string
			for _, row := range rows {
				require.Equal(t, failedUnitsLabel, row.Label.String())
				values = append(values, row.Value.String())
			}
			assert.Equal(t, tt.expected, values)
		})
	}
}
