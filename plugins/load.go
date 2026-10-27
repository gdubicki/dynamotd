package plugins

import (
	"fmt"
	"runtime"

	"github.com/gdubicki/dynamotd/dynamotd"
	"github.com/shirou/gopsutil/v4/load"
)

func init() {
	dynamotd.Register(dynamotd.RowType{Label: loadLabel, Generate: dynamotd.SingleRow(Load)}, "load")
}

const loadLabel = "CPU Load (1/5/15)"

func Load() dynamotd.Row {
	// TODO: make below configurable
	// (thresholds are per core)
	var warningThreshold = 0.9
	var criticalThreshold = 1.1

	cores := float64(runtime.NumCPU())

	avgStat, err := load.Avg()
	if err != nil {
		return dynamotd.ErrorRow(loadLabel, fmt.Errorf("error getting load info: %w", err))
	}

	coresDescription, err := getCoresDescription()
	if err != nil {
		return dynamotd.ErrorRow(loadLabel, err)
	}

	loadString := func(load float64) dynamotd.ColorString {
		return dynamotd.ColorString{
			Color: dynamotd.ThresholdColor(load, warningThreshold*cores, criticalThreshold*cores),
			Text:  fmt.Sprintf("%.2f", load),
		}
	}

	value := []dynamotd.ColorString{
		loadString(avgStat.Load1),
		dynamotd.ValueDescription(" / "),
		loadString(avgStat.Load5),
		dynamotd.ValueDescription(" / "),
		loadString(avgStat.Load15),
		dynamotd.ValueDescription(" with "),
	}
	value = append(value, coresDescription...)

	return dynamotd.Row{
		Label: dynamotd.SingleColorLabel(loadLabel),
		Value: dynamotd.ToColorText(value...),
	}
}
