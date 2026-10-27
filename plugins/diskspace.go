package plugins

import (
	"fmt"

	"github.com/Tonyfilla/go-humanize"
	"github.com/gdubicki/dynamotd/dynamotd"
	"github.com/shirou/gopsutil/v4/disk"
)

func init() {
	dynamotd.Register("Disk space (/)", dynamotd.Thresholds{Warning: 75, Critical: 90},
		dynamotd.SingleRow(func(thresholds dynamotd.Thresholds) dynamotd.Row { return DiskSpace("/", thresholds) }),
		"diskspace")
}

// DiskSpace shows the disk space usage of the filesystem with the given path, with the thresholds in %.
func DiskSpace(path string, thresholds dynamotd.Thresholds) dynamotd.Row {
	label := fmt.Sprintf("Disk space (%s)", path)

	du, err := disk.Usage(path)
	if err != nil {
		return dynamotd.ErrorRow(label, fmt.Errorf("error getting disk info: %w", err))
	}

	percentageUsed := du.UsedPercent

	color := thresholds.Color(percentageUsed)

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
