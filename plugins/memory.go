package plugins

import (
	"context"
	"fmt"

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

const memoryLabel = "Memory"

// for mocking in tests
var getVirtualMemory = mem.VirtualMemoryWithContext
var getSwap = mem.SwapMemoryWithContext

// Memory shows the RAM usage, and the swap usage if there is any swap.
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

	return []dynamotd.Row{{
		Label: dynamotd.SingleColorLabel(memoryLabel),
		Value: dynamotd.ToColorText(value...),
	}}
}
