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
	memoryAvailableBytes := float64(memoryStat.Total) - float64(memoryStat.Used)

	if memoryAvailableBytes <= criticalThreshold * memoryTotalBytes {
		color = valueCriticalColor
	} else if memoryAvailableBytes <= warningThreshold * memoryTotalBytes {
		color = valueWarningColor
	} else {
		color = valueOkColor
	}

	memoryTotalGB := memoryTotalBytes / 1024 / 1024 / 1024
	memoryAvailableGB := memoryAvailableBytes / 1024 / 1024 / 1024

	return Row{
		singleColorLabel("Memory"),
		toColorText(
			ColorString{color, fmt.Sprintf("%.2f GB", memoryAvailableGB)},
			valueDescription(" (available) / "),
			valueNeutral(fmt.Sprintf("%.2f GB", memoryTotalGB)),
			valueDescription(" (total)"),
		),
	}
}
