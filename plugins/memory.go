package plugins

import (
	"fmt"
	. "github.com/gdubicki/dynamotd/dynamotd"
	humanize "github.com/dustin/go-humanize"
	memoryLib "github.com/mackerelio/go-osstat/memory"
)

func Memory() Row {
	var warningThreshold = 90.0
	var criticalThreshold = 95.0

	var color Color

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
		Label: SingleColorLabel("Memory"),
		Value: ToColorText(
			ColorString{Color: color, Text: fmt.Sprintf("%s", humanize.Bytes(memoryUsedBytes))},
			ValueDescription(" of "),
			ValueNeutral(fmt.Sprintf("%s", humanize.Bytes(memoryTotalBytes))),
			ValueDescription(" RAM used ("),
			ColorString{Color: color, Text: fmt.Sprintf("%0.2f%%", percentage)},
			ValueDescription(")"),
		),
	}
}
