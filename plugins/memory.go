package plugins

import (
	"fmt"

	"github.com/Tonyfilla/go-humanize"
	"github.com/gdubicki/dynamotd/dynamotd"
	"github.com/shirou/gopsutil/v4/mem"
)

func init() {
	dynamotd.Register(Memory, "memory")
}

const memoryLabel = "RAM"

func Memory() dynamotd.Row {
	var warningThreshold = 90.0
	var criticalThreshold = 95.0

	memoryStat, err := mem.VirtualMemory()
	if err != nil {
		return dynamotd.ErrorRow(memoryLabel, fmt.Errorf("error getting memory info: %w", err))
	}
	memoryTotalBytes := memoryStat.Total
	// memory that is really used, so without caches and buffers that can be freed
	memoryUsedBytes := memoryStat.Total - memoryStat.Available

	percentage := float64(memoryUsedBytes) / float64(memoryTotalBytes) * 100

	color := dynamotd.ThresholdColor(percentage, warningThreshold, criticalThreshold)

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
