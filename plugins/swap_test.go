package plugins

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/gdubicki/dynamotd/dynamotd"
	"github.com/shirou/gopsutil/v4/mem"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func mockSwap(t *testing.T, swap mem.SwapMemoryStat, vmstat string) {
	origSwap, origPath := getSwap, vmstatPath
	t.Cleanup(func() { getSwap, vmstatPath = origSwap, origPath })

	getSwap = func(context.Context) (*mem.SwapMemoryStat, error) { return &swap, nil }
	vmstatPath = filepath.Join(t.TempDir(), "vmstat")
	if vmstat != "" {
		require.NoError(t, os.WriteFile(vmstatPath, []byte(vmstat), 0o644))
	}
}

func TestSwap(t *testing.T) {
	const gib = 1024 * 1024 * 1024
	tests := []struct {
		name     string
		swap     mem.SwapMemoryStat
		vmstat   string
		expected string
		color    string
	}{
		{"no swap, no vmstat (macOS)", mem.SwapMemoryStat{}, "", "no swap", ""},
		{"no OOM kills", mem.SwapMemoryStat{Total: 2 * gib, Used: gib / 2, UsedPercent: 25},
			"nr_free_pages 1\noom_kill 0\n", "512 MiB of 2 GiB swap used (25%)", "ok"},
		{"OOM kills", mem.SwapMemoryStat{Total: 2 * gib, Used: gib + gib/2, UsedPercent: 75},
			"oom_kill 3\nnr_foo 5\n", "1.5 GiB of 2 GiB swap used (75%), 3 OOM kills since boot", "warning"},
		{"one OOM kill, no swap", mem.SwapMemoryStat{}, "oom_kill 1\n", "no swap, 1 OOM kill since boot", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSwap(t, tt.swap, tt.vmstat)

			rows := Swap(context.Background(), dynamotd.Thresholds{Warning: 50, Critical: 80})

			require.Len(t, rows, 1)
			assert.Equal(t, tt.expected, rows[0].Value.String())
			switch tt.color {
			case "ok":
				assert.Equal(t, dynamotd.ValueOkColor, rows[0].Value.Parts()[0].Color)
			case "warning":
				assert.Equal(t, dynamotd.ValueWarningColor, rows[0].Value.Parts()[0].Color)
			}
		})
	}
}
