package plugins

import (
	. "github.com/gdubicki/dynamotd/dynamotd"
	"fmt"
	memoryLib "github.com/mackerelio/go-osstat/memory"
)

func Memory() Row {
	var warningThreshold = 0.1
	var criticalThreshold = 0.05

	var color Color

	memoryStat, err := memoryLib.Get()
	if err != nil {
		panic(fmt.Errorf("error getting memory info"))
	}
	memoryTotalBytes := float64(memoryStat.Total)
	memoryAvailableBytes := float64(memoryStat.Total) - float64(memoryStat.Used)

	if memoryAvailableBytes <= criticalThreshold * memoryTotalBytes {
		color = ValueCriticalColor
	} else if memoryAvailableBytes <= warningThreshold * memoryTotalBytes {
		color = ValueWarningColor
	} else {
		color = ValueOkColor
	}

	memoryTotalGB := memoryTotalBytes / 1024 / 1024 / 1024
	memoryAvailableGB := memoryAvailableBytes / 1024 / 1024 / 1024

	return Row {
		Label: SingleColorLabel("Memory"),
		Value: ToColorText(
			ColorString{Color: color, Text: fmt.Sprintf("%.2f GB", memoryAvailableGB)},
			ValueDescription(" (available) / "),
			ValueNeutral(fmt.Sprintf("%.2f GB", memoryTotalGB)),
			ValueDescription(" (total)"),
		),
	}
}
