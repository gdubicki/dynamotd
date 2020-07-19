package main

import (
	"fmt"
	"github.com/shirou/gopsutil/host"
)


func uptime() Row {

	color, uptimeString := getUptimeColorAndString()

	return Row{
		singleColorLabel("Uptime"),
		toColorText(
			ColorString{color, uptimeString},
		),
	}
}

func getUptimeColorAndString() (Color, string) {
	var color Color

	uptime,_ := host.Uptime()

	var secondsInAYear uint64 = 60 * 60 * 24 * 365
	years := uptime / secondsInAYear

	var secondsInAMonth uint64 = 60 * 60 * 24 * 30
	months := (uptime - (years * secondsInAYear)) / secondsInAMonth

	var secondsInADay uint64 = 60 * 60 * 24
	days := (uptime - (years * secondsInAYear) - (months * secondsInAMonth)) / secondsInADay

	var secondsInAnHour uint64 = 60 * 60
	hours := (uptime - (years * secondsInAYear) - (months * secondsInAMonth) - (days * secondsInADay)) / secondsInAnHour

	var secondsInAMinute uint64 = 60
	minutes := (uptime - (years * secondsInAYear) - (months * secondsInAMonth) - (days * secondsInADay) - (hours * secondsInAnHour)) / secondsInAMinute

	// TODO: make this configurable
	if years >= 1 {
		color = valueCriticalColor
	} else if months >= 1 {
		color = valueWarningColor
	} else {
		color = valueOkColor
	}

	uptimeString := ""
	if years >= 1 {
		uptimeString += fmt.Sprintf("%d year(s), ", years)
	}
	if months >= 1 {
		uptimeString += fmt.Sprintf("%d month(s), ", months)
	}
	if days >= 1 {
		uptimeString += fmt.Sprintf("%d day(s), ", days)
	}
	if hours >= 1 {
		uptimeString += fmt.Sprintf("%d hour(s), ", hours)
	}
	uptimeString += fmt.Sprintf("%d minute(s)", minutes)

	return color, uptimeString
}
