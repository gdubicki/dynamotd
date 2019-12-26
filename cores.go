package main

import (
	"fmt"
	"github.com/shirou/gopsutil/cpu"
	"runtime"
)


func cores() Row {

	cores := runtime.NumCPU()

	cpuStat, err := cpu.Info()
	if err != nil {
		panic(fmt.Errorf("error getting CPU info"))
	}
	mhz := int(cpuStat[0].Mhz)

	return Row{
		singleColorLabelText("Core(s)"),
		singleColorValueText(fmt.Sprintf("%d core(s) at %d MHz", cores, mhz)),
	}
}
