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

const gib = 1024 * 1024 * 1024

func mockMemory(t *testing.T, memory mem.VirtualMemoryStat, swap mem.SwapMemoryStat, vmstat string) {
	origMemory, origSwap, origPath := getVirtualMemory, getSwap, vmstatPath
	t.Cleanup(func() { getVirtualMemory, getSwap, vmstatPath = origMemory, origSwap, origPath })

	getVirtualMemory = func(context.Context) (*mem.VirtualMemoryStat, error) { return &memory, nil }
	getSwap = func(context.Context) (*mem.SwapMemoryStat, error) { return &swap, nil }
	vmstatPath = filepath.Join(t.TempDir(), "vmstat")
	if vmstat != "" {
		require.NoError(t, os.WriteFile(vmstatPath, []byte(vmstat), 0o644))
	}
}

var defaultMemorySettings = memorySettings{
	Thresholds: dynamotd.Thresholds{Warning: 90, Critical: 95},
	Swap:       dynamotd.Thresholds{Warning: 50, Critical: 80},
}

func TestMemory(t *testing.T) {
	ram := mem.VirtualMemoryStat{Total: 16 * gib, Available: 12 * gib}

	tests := []struct {
		name     string
		swap     mem.SwapMemoryStat
		vmstat   string
		expected string
	}{
		{"no swap, no vmstat (macOS)", mem.SwapMemoryStat{}, "",
			"4 GiB of 16 GiB RAM used (25%)"},
		{"swap, no OOM kills", mem.SwapMemoryStat{Total: 2 * gib, Used: gib / 2, UsedPercent: 25}, "nr_free_pages 1\noom_kill 0\n",
			"4 GiB of 16 GiB RAM used (25%), 512 MiB of 2 GiB swap (25%)"},
		{"swap and OOM kills", mem.SwapMemoryStat{Total: 2 * gib, Used: gib + gib/2, UsedPercent: 75}, "oom_kill 3\nnr_foo 5\n",
			"4 GiB of 16 GiB RAM used (25%), 1.5 GiB of 2 GiB swap (75%), 3 OOM kills since boot"},
		{"no swap, one OOM kill", mem.SwapMemoryStat{}, "oom_kill 1\n",
			"4 GiB of 16 GiB RAM used (25%), 1 OOM kill since boot"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockMemory(t, ram, tt.swap, tt.vmstat)

			rows := Memory(context.Background(), defaultMemorySettings)

			require.Len(t, rows, 1)
			assert.Equal(t, tt.expected, rows[0].Value.String())
		})
	}
}

func TestMemoryColors(t *testing.T) {
	mockMemory(t, mem.VirtualMemoryStat{Total: 100 * gib, Available: 8 * gib},
		mem.SwapMemoryStat{Total: 2 * gib, Used: gib + gib/2, UsedPercent: 75}, "")

	parts := Memory(context.Background(), defaultMemorySettings)[0].Value.Parts()

	assert.Equal(t, dynamotd.ValueWarningColor, parts[0].Color, "RAM at 92%")
	assert.Equal(t, dynamotd.ValueWarningColor, parts[7].Color, "swap at 75%")
}

func TestMemorySettingsValidation(t *testing.T) {
	settings := defaultMemorySettings
	settings.Swap = dynamotd.Thresholds{Warning: 90, Critical: 80}
	assert.ErrorContains(t, settings.Validate(), "swap: warning threshold")
}
