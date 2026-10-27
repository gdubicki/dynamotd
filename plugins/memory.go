package plugins

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/Tonyfilla/go-humanize"
	"github.com/gdubicki/dynamotd/dynamotd"
	"github.com/shirou/gopsutil/v4/mem"
)

type memorySettings struct {
	// thresholds for the RAM used, in %
	dynamotd.Thresholds `yaml:",inline"`
	// thresholds for the swap used, in %
	Swap dynamotd.Thresholds `yaml:"swap"`
}

func (settings memorySettings) Validate() error {
	if err := settings.Thresholds.Validate(); err != nil {
		return err
	}
	if err := settings.Swap.Validate(); err != nil {
		return fmt.Errorf("swap: %w", err)
	}
	return nil
}

func init() {
	dynamotd.Register(memoryLabel, memorySettings{
		Thresholds: dynamotd.Thresholds{Warning: 90, Critical: 95},
		Swap:       dynamotd.Thresholds{Warning: 50, Critical: 80},
	}, Memory, "memory")
}

const memoryLabel = "RAM"

// for mocking in tests
var getVirtualMemory = mem.VirtualMemoryWithContext
var getSwap = mem.SwapMemoryWithContext
var vmstatPath = "/proc/vmstat"

// Memory shows the RAM usage, the swap usage if there is any swap, and the number of processes killed
// by the OOM killer since boot, if there were any.
func Memory(ctx context.Context, settings memorySettings) []dynamotd.Row {
	memoryStat, err := getVirtualMemory(ctx)
	if err != nil {
		return []dynamotd.Row{dynamotd.ErrorRow(memoryLabel, fmt.Errorf("error getting memory info: %w", err))}
	}
	memoryTotalBytes := memoryStat.Total
	// memory that is really used, so without caches and buffers that can be freed
	memoryUsedBytes := memoryStat.Total - memoryStat.Available

	percentage := float64(memoryUsedBytes) / float64(memoryTotalBytes) * 100

	color := settings.Color(percentage)

	value := []dynamotd.ColorString{
		{Color: color, Text: humanize.IBytesCustomCeil(memoryUsedBytes, 2)},
		dynamotd.ValueDescription(" of "),
		dynamotd.ValueNeutral(humanize.IBytesCustomCeil(memoryTotalBytes, 2)),
		dynamotd.ValueDescription(" RAM used ("),
		{Color: color, Text: fmt.Sprintf("%0.0f%%", percentage)},
		dynamotd.ValueDescription(")"),
	}

	if swap, err := getSwap(ctx); err == nil && swap.Total > 0 {
		swapColor := settings.Swap.Color(swap.UsedPercent)
		value = append(value,
			dynamotd.ValueDescription(", "),
			dynamotd.ColorString{Color: swapColor, Text: humanize.IBytesCustomCeil(swap.Used, 2)},
			dynamotd.ValueDescription(" of "),
			dynamotd.ValueNeutral(humanize.IBytesCustomCeil(swap.Total, 2)),
			dynamotd.ValueDescription(" swap ("),
			dynamotd.ColorString{Color: swapColor, Text: fmt.Sprintf("%0.0f%%", swap.UsedPercent)},
			dynamotd.ValueDescription(")"),
		)
	}

	// it's not available on macOS and old Linux kernels
	if oomKills, err := getOOMKills(); err == nil && oomKills > 0 {
		unit := "OOM kill"
		if oomKills > 1 {
			unit += "s"
		}
		value = append(value,
			dynamotd.ValueDescription(", "),
			dynamotd.ValueCritical(fmt.Sprintf("%d %s", oomKills, unit)),
			dynamotd.ValueDescription(" since boot"),
		)
	}

	return []dynamotd.Row{{
		Label: dynamotd.SingleColorLabel(memoryLabel),
		Value: dynamotd.ToColorText(value...),
	}}
}

// getOOMKills returns the number of processes killed by the OOM killer since boot.
func getOOMKills() (uint64, error) {
	file, err := os.Open(vmstatPath)
	if err != nil {
		return 0, err
	}
	defer func() { _ = file.Close() }()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		if value, found := strings.CutPrefix(scanner.Text(), "oom_kill "); found {
			return strconv.ParseUint(value, 10, 64)
		}
	}
	if err := scanner.Err(); err != nil {
		return 0, err
	}
	return 0, fmt.Errorf("no oom_kill in %s", vmstatPath)
}
