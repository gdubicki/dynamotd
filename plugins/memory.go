package plugins

import (
	"fmt"

	"github.com/Tonyfilla/go-humanize"
	"github.com/gdubicki/dynamotd/dynamotd"
	"github.com/shirou/gopsutil/v4/mem"
)

func init() {
	dynamotd.Register(memoryLabel, dynamotd.Thresholds{Warning: 90, Critical: 95}, dynamotd.SingleRow(Memory), "memory")
}

const memoryLabel = "RAM"

// Memory shows the RAM usage, with the thresholds in %.
func Memory(thresholds dynamotd.Thresholds) dynamotd.Row {
	memoryStat, err := mem.VirtualMemory()
	if err != nil {
		return dynamotd.ErrorRow(memoryLabel, fmt.Errorf("error getting memory info: %w", err))
	}
	memoryTotalBytes := memoryStat.Total
	// memory that is really used, so without caches and buffers that can be freed
	memoryUsedBytes := memoryStat.Total - memoryStat.Available

	percentage := float64(memoryUsedBytes) / float64(memoryTotalBytes) * 100

	color := thresholds.Color(percentage)

	return dynamotd.Row{
		Label: dynamotd.SingleColorLabel(memoryLabel),
		Value: dynamotd.ToColorText(
			dynamotd.ColorString{Color: color, Text: humanize.IBytesCustomCeil(memoryUsedBytes, 2)},
			dynamotd.ValueDescription(" of "),
			dynamotd.ValueNeutral(humanize.IBytesCustomCeil(memoryTotalBytes, 2)),
			dynamotd.ValueDescription(" RAM used ("),
			dynamotd.ColorString{Color: color, Text: fmt.Sprintf("%0.0f%%", percentage)},
			dynamotd.ValueDescription(")"),
		),
	}
}
