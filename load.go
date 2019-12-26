package main

import (
	"fmt"
	linuxproc "github.com/c9s/goprocinfo/linux"
	"runtime"
)

func load() Row {
	var warningThreshold = 0.9
	var criticalThreshold = 1.1

	var cores = float64(runtime.NumCPU())

	var load1, load5, load15 float64

	stat, err := linuxproc.ReadLoadAvg("/proc/stat")
	if err == nil {
		load1 = stat.Last1Min
		load5 = stat.Last5Min
		load15 = stat.Last15Min
	} else {
		fmt.Printf("Can't read /proc, faking load values\n")
		load1 = 1.0
		load5 = 5.0
		load15 = 20.0
	}

	var load1color, load5color, load15color Color

	if load1 >= warningThreshold*cores {
		load1color = keyValueWarningColor
	}
	if load5 >= warningThreshold*cores {
		load5color = keyValueWarningColor
	}
	if load15 >= warningThreshold*cores {
		load15color = keyValueWarningColor
	}

	if load1 >= criticalThreshold*cores {
		load1color = keyValueCriticalColor
	}
	if load5 >= criticalThreshold*cores {
		load5color = keyValueCriticalColor
	}
	if load15 >= criticalThreshold*cores {
		load15color = keyValueCriticalColor
	}

	// text format like: 0.12, 0.4, 0.5 (1 / 5 / 15), with colors
	return Row{
		singleColorLabelText("Load"),
		toColorText(
			ColorString{load1color, fmt.Sprintf("%f", load1)},
			value(", "),
			ColorString{load5color, fmt.Sprintf("%f", load5)},
			value(", "),
			ColorString{load15color, fmt.Sprintf("%f", load15)},
			value(" (1 / 5 / 15)"),
		),
	}
}
