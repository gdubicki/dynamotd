package plugins

import (
	"fmt"
	. "github.com/gdubicki/dynamotd/dynamotd"
	"github.com/shirou/gopsutil/v3/cpu"
	"runtime"
)

func Cores() Row {

	cores := runtime.NumCPU()

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
		Label: SingleColorLabel("CPU"),
		Value: ToColorText(
			ValueNeutral(fmt.Sprintf("%d", cores)),
			ValueDescription(fmt.Sprintf(" %s at ", unit)),
			ValueNeutral(fmt.Sprintf("%d MHz", mhz)),
		),
	}
}
