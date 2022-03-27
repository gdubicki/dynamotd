package plugins

import (
	"fmt"
	"github.com/Tonyfilla/go-humanize"
	"github.com/fatih/color"
	. "github.com/gdubicki/dynamotd/dynamotd"
	memoryLib "github.com/mackerelio/go-osstat/memory"
)

func Memory() Row {
	var warningThreshold = 90.0
	var criticalThreshold = 95.0

	var color color.Attribute

	memoryStat, err := memoryLib.Get()
	if err != nil {
		panic(fmt.Errorf("error getting memory info"))
	}
	memoryTotalBytes := memoryStat.Total
	memoryUsedBytes := memoryStat.Used

	percentage := float64(memoryUsedBytes) / float64(memoryTotalBytes) * 100

	if percentage >= criticalThreshold {
		color = ValueCriticalColor
	} else if percentage >= warningThreshold {
		color = ValueWarningColor
	} else {
		color = ValueOkColor
	}

	return Row{
		Label: SingleColorLabel("RAM"),
		Value: ToColorText(
			ColorString{Color: color, Text: fmt.Sprintf("%s", humanize.IBytesCustomCeil(memoryUsedBytes, 2))},
			ValueDescription(" of "),
			ValueNeutral(fmt.Sprintf("%s", humanize.IBytesCustomCeil(memoryTotalBytes, 2))),
			ValueDescription(" RAM used ("),
			ColorString{Color: color, Text: fmt.Sprintf("%0.0f%%", percentage)},
			ValueDescription(")"),
		),
	}
}
