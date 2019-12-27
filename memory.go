package main

import (
	"fmt"
	memoryLib "github.com/mackerelio/go-osstat/memory"
)

func memory() Row {
	var warningThreshold = 0.1
	var criticalThreshold = 0.05

	var color Color

	memoryStat, err := memoryLib.Get()
	if err != nil {
		panic(fmt.Errorf("error getting memory info"))
	}
	memoryTotalBytes := float64(memoryStat.Total)
	memoryFreeBytes := float64(memoryStat.Free)

	if memoryFreeBytes <= criticalThreshold*memoryTotalBytes {
		color = keyValueCriticalColor
	} else if memoryFreeBytes <= warningThreshold*memoryTotalBytes {
		color = keyValueWarningColor
	} else {
		color = keyValueOkColor
	}

	memoryTotalGB := memoryTotalBytes / 1024 / 1024 / 1024
	memoryFreeGB := memoryFreeBytes / 1024 / 1024 / 1024

	return Row{
		singleColorLabelText("Memory"),
		toColorText(
			ColorString{color, fmt.Sprintf("%.2f GB", memoryFreeGB)},
			value(" (free) / "),
			value(fmt.Sprintf("%.2f GB", memoryTotalGB)),
			value(" (total)"),
		),
	}
}
