package main

import (
	"fmt"
	"github.com/shirou/gopsutil/host"
)


func uptime() Row {

	color, uptimeString := getUptimeColorAndString()

	return Row{
		singleColorLabelText("Uptime"),
		toColorText(
			ColorString{color, uptimeString},
		),
	}
}

func getUptimeColorAndString() (Color, string) {
	var color Color

	uptime,_ := host.Uptime()

	years := uptime / (60 * 60 * 24 * 365)
	months := uptime / (60 * 60 * 24 * 30)
	days := uptime / (60 * 60 * 24)
	hours := (uptime - (days * 60 * 60 * 24)) / (60 * 60)
	minutes := ((uptime - (days * 60 * 60 * 24))  -  (hours * 60 * 60)) / 60

	if years >= 1 {
		color = keyValueCriticalColor
	} else if months >= 1 {
		color = keyValueWarningColor
	} else {
		color = keyValueOkColor
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
