package plugins

import (
	"fmt"
	"github.com/fatih/color"
	. "github.com/gdubicki/dynamotd/dynamotd"
	"github.com/shirou/gopsutil/host"
)

func Uptime() Row {

	color, uptimeString := getUptimeColorAndString()

	return Row{
		Label: SingleColorLabel("Uptime"),
		Value: ToColorText(
			ColorString{Color: color, Text: uptimeString},
		),
	}
}

func getUptimeColorAndString() (color.Attribute, string) {
	uptimeSeconds, _ := host.Uptime()

	uptimeColor := GetUptimeColor(uptimeSeconds)
	uptimeString := GetUptimeString(uptimeSeconds)

	return uptimeColor, uptimeString
}

func GetUptimeColor(uptimeSeconds uint64) color.Attribute {
	uptimeValues := getUptimeValues(uptimeSeconds)

	// TODO: make this configurable
	if uptimeValues["year"] >= 1 {
		return ValueCriticalColor
	} else if uptimeValues["month"] >= 3 {
		return ValueWarningColor
	} else {
		return ValueOkColor
	}
}

func GetUptimeString(uptimeSeconds uint64) string {
	if uptimeSeconds < 60 {
		return "less than 1 minute"
	}

	uptimeValues := getUptimeValues(uptimeSeconds)
	var uptimeUnits = []string{"year", "month", "day", "hour", "minute"}

	uptimeString := ""
	first := true

	for unitNo := 0; unitNo < 5; unitNo++ {

		unitSingular := uptimeUnits[unitNo]
		unitPlural := uptimeUnits[unitNo] + "s"

		if uptimeValues[unitPlural] > 0 {
			if !first {
				uptimeString += ", "
			}

			if uptimeValues[unitPlural] == 1 {
				uptimeString += fmt.Sprintf("%d %s", uptimeValues[unitPlural], unitSingular)
			} else { // >= 2
				uptimeString += fmt.Sprintf("%d %s", uptimeValues[unitPlural], unitPlural)
			}

			first = false
		}
	}

	return uptimeString
}

func getUptimeValues(uptimeSeconds uint64) map[string]uint64 {
	var uptimeValues = make(map[string]uint64)

	var secondsInAYear uint64 = 60 * 60 * 24 * 365
	uptimeValues["years"] = uptimeSeconds / secondsInAYear
	remainingUptime := uptimeSeconds - (uptimeValues["years"] * secondsInAYear)

	var secondsInAMonth uint64 = 60 * 60 * 24 * 30
	uptimeValues["months"] = remainingUptime / secondsInAMonth
	remainingUptime = remainingUptime - (uptimeValues["months"] * secondsInAMonth)

	var secondsInADay uint64 = 60 * 60 * 24
	uptimeValues["days"] = remainingUptime / secondsInADay
	remainingUptime = remainingUptime - (uptimeValues["days"] * secondsInADay)

	var secondsInAnHour uint64 = 60 * 60
	uptimeValues["hours"] = remainingUptime / secondsInAnHour
	remainingUptime = remainingUptime - (uptimeValues["hours"] * secondsInAnHour)

	var secondsInAMinute uint64 = 60
	uptimeValues["minutes"] = remainingUptime / secondsInAMinute
	remainingUptime = remainingUptime - (uptimeValues["hours"] * secondsInAnHour)

	uptimeValues["seconds"] = remainingUptime

	return uptimeValues
}
