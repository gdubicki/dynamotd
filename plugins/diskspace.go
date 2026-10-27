package plugins

import (
	"fmt"

	"github.com/Tonyfilla/go-humanize"
	"github.com/gdubicki/dynamotd/dynamotd"
	"github.com/shirou/gopsutil/v4/disk"
)

func init() {
	dynamotd.Register(dynamotd.RowType{
		Label:    "Disk space (/)",
		Generate: dynamotd.SingleRow(func() dynamotd.Row { return DiskSpace("/") }),
	}, "diskspace")
}

func DiskSpace(path string) dynamotd.Row {
	var warningThreshold = 75.0
	var criticalThreshold = 90.0

	label := fmt.Sprintf("Disk space (%s)", path)

	du, err := disk.Usage(path)
	if err != nil {
		return dynamotd.ErrorRow(label, fmt.Errorf("error getting disk info: %w", err))
	}

	percentageUsed := du.UsedPercent

	color := dynamotd.ThresholdColor(percentageUsed, warningThreshold, criticalThreshold)

	return dynamotd.Row{
		Label: dynamotd.SingleColorLabel(label),
		Value: dynamotd.ToColorText(
			dynamotd.ColorString{Color: color, Text: humanize.IBytesCustomCeil(du.Used, 2)},
			dynamotd.ValueDescription(" of "),
			dynamotd.ValueNeutral(humanize.IBytesCustomCeil(du.Total, 0)),
			dynamotd.ValueDescription(" disk space used ("),
			dynamotd.ColorString{Color: color, Text: fmt.Sprintf("%0.0f%%", percentageUsed)},
			dynamotd.ValueDescription(")"),
		),
	}
}
