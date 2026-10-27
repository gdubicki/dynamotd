package plugins

import (
	"context"
	"testing"

	"github.com/gdubicki/dynamotd/dynamotd"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClock(t *testing.T) {
	orig := clockSynchronized
	t.Cleanup(func() { clockSynchronized = orig })

	tests := []struct {
		name         string
		synchronized bool
		ok           bool
		expected     string
	}{
		{"synchronized", true, true, ""},
		{"unknown", false, false, ""},
		{"not synchronized", false, true, "not synchronized - check NTP"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clockSynchronized = func() (bool, bool) { return tt.synchronized, tt.ok }

			rows := Clock(context.Background(), dynamotd.NoSettings{})

			if tt.expected == "" {
				assert.Empty(t, rows)
				return
			}
			require.Len(t, rows, 1)
			assert.Equal(t, tt.expected, rows[0].Value.String())
		})
	}
}

// the real check should work, whatever its result is
func TestIsClockSynchronizedDoesNotPanic(t *testing.T) {
	synchronized, ok := isClockSynchronized()
	t.Logf("synchronized: %v, ok: %v", synchronized, ok)
}
