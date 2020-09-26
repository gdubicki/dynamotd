package plugins

import (
	"fmt"
	. "github.com/gdubicki/dynamotd/dynamotd"
	"github.com/shirou/gopsutil/cpu"
	"runtime"
)


func Cores() Row {

	cores := runtime.NumCPU()

	cpuStat, err := cpu.Info()
	if err != nil {
		panic(fmt.Errorf("error getting CPU info"))
	}
	mhz := int(cpuStat[0].Mhz)

	return Row{
		SingleColorLabel("Core(s)"),
		ToColorText(
			ValueNeutral(fmt.Sprintf("%d", cores)),
			ValueDescription(" core(s) at "),
			ValueNeutral(fmt.Sprintf("%d MHz", mhz)),
		),
	}
}
