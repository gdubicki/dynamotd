package plugins

import (
	"fmt"
	"runtime"

	"github.com/gdubicki/dynamotd/dynamotd"
	"github.com/shirou/gopsutil/v4/cpu"
)

func init() {
	dynamotd.Register(Cores, "cores")
}

const coresLabel = "CPU"

func Cores() dynamotd.Row {
	coresDescription, err := getCoresDescription()
	if err != nil {
		return dynamotd.ErrorRow(coresLabel, err)
	}

	return dynamotd.Row{
		Label: dynamotd.SingleColorLabel(coresLabel),
		Value: dynamotd.ToColorText(coresDescription...),
	}
}

// getCoresDescription returns "<n> core(s) at <n> MHz"
func getCoresDescription() ([]dynamotd.ColorString, error) {
	cores := runtime.NumCPU()

	cpuStat, err := cpu.Info()
	if err != nil {
		return nil, fmt.Errorf("error getting CPU info: %w", err)
	}
	if len(cpuStat) == 0 {
		return nil, fmt.Errorf("no CPU info")
	}
	mhz := int(cpuStat[0].Mhz)

	unit := "core"
	if cores >= 2 {
		unit += "s"
	}

	return []dynamotd.ColorString{
		dynamotd.ValueNeutral(fmt.Sprintf("%d", cores)),
		dynamotd.ValueDescription(fmt.Sprintf(" %s at ", unit)),
		dynamotd.ValueNeutral(fmt.Sprintf("%d MHz", mhz)),
	}, nil
}
