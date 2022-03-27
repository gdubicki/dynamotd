package plugins

import (
	"fmt"
	. "github.com/gdubicki/dynamotd/dynamotd"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/load"
	"runtime"
)

func Load() Row {
	// TODO: make below configurable
	var warningThreshold = 0.9
	var criticalThreshold = 1.1

	var cores = float64(runtime.NumCPU())

	var load1, load5, load15 float64

	avgStat, err := load.Avg()
	if err != nil {
		panic(fmt.Errorf("error getting load info"))
	}

	load1 = avgStat.Load1
	load5 = avgStat.Load5
	load15 = avgStat.Load15

	load1color := ValueOkColor
	load5color := ValueOkColor
	load15color := ValueOkColor

	if load1 >= criticalThreshold*cores {
		load1color = ValueCriticalColor
	} else if load1 >= warningThreshold*cores {
		load1color = ValueWarningColor
	}

	if load5 >= criticalThreshold*cores {
		load5color = ValueCriticalColor
	} else if load5 >= warningThreshold*cores {
		load5color = ValueWarningColor
	}

	if load15 >= criticalThreshold*cores {
		load15color = ValueCriticalColor
	} else if load15 >= warningThreshold*cores {
		load15color = ValueWarningColor
	}

	cpuStat, err := cpu.Info()
	if err != nil {
		panic(fmt.Errorf("error getting CPU info"))
	}
	mhz := int(cpuStat[0].Mhz)

	unit := "core"
	if cores >= 2 {
		unit = unit + "s"
	}

	return Row{
		Label: SingleColorLabel("CPU Load (1/5/15)"),
		Value: ToColorText(
			ColorString{Color: load1color, Text: fmt.Sprintf("%.2f", load1)},
			ValueDescription(" / "),
			ColorString{Color: load5color, Text: fmt.Sprintf("%.2f", load5)},
			ValueDescription(" / "),
			ColorString{Color: load15color, Text: fmt.Sprintf("%.2f", load15)},
			ValueDescription(" with "),
			ValueNeutral(fmt.Sprintf("%d", int(cores))),
			ValueDescription(fmt.Sprintf(" %s at ", unit)),
			ValueNeutral(fmt.Sprintf("%d MHz", mhz)),
		),
	}
}
