package plugins

import (
	"fmt"
	"github.com/Tonyfilla/go-humanize"
	"github.com/fatih/color"
	. "github.com/gdubicki/dynamotd/dynamotd"
	"github.com/shirou/gopsutil/v3/disk"
)

func DiskSpace(path string) Row {
	var warningThreshold = 75.0
	var criticalThreshold = 90.0

	var color color.Attribute

	du, err := disk.Usage(path)
	if err != nil {
		panic(fmt.Errorf("error getting disk info"))
	}

	percentageUsed := du.UsedPercent

	if percentageUsed >= criticalThreshold {
		color = ValueCriticalColor
	} else if percentageUsed >= warningThreshold {
		color = ValueWarningColor
	} else {
		color = ValueOkColor
	}

	return Row{
		Label: SingleColorLabel(fmt.Sprintf("Disk space (%s)", path)),
		Value: ToColorText(
			ColorString{Color: color, Text: fmt.Sprintf("%s", humanize.IBytesCustomCeil(du.Used, 2))},
			ValueDescription(" of "),
			ValueNeutral(fmt.Sprintf("%s", humanize.IBytesCustomCeil(du.Total, 0))),
			ValueDescription(" disk space used ("),
			ColorString{Color: color, Text: fmt.Sprintf("%0.0f%%", percentageUsed)},
			ValueDescription(")"),
		),
	}
}
