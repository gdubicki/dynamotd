package plugins

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
	"github.com/gdubicki/dynamotd/dynamotd"
	"github.com/shirou/gopsutil/v4/host"
)

func init() {
	dynamotd.Register(uptimeLabel, dynamotd.Thresholds{Warning: 90, Critical: 365}, dynamotd.SingleRow(Uptime), "uptime")
}

const uptimeLabel = "Uptime"

// Uptime shows the time since the last boot, with the thresholds in days, as a long uptime
// usually means that the latest kernel security fixes are not applied.
func Uptime(thresholds dynamotd.Thresholds) dynamotd.Row {
	uptimeSeconds, err := host.Uptime()
	if err != nil {
		return dynamotd.ErrorRow(uptimeLabel, fmt.Errorf("error getting uptime: %w", err))
	}

	return dynamotd.Row{
		Label: dynamotd.SingleColorLabel(uptimeLabel),
		Value: dynamotd.ToColorText(
			dynamotd.ColorString{Color: GetUptimeColor(uptimeSeconds, thresholds), Text: GetUptimeString(uptimeSeconds)},
		),
	}
}

func GetUptimeColor(uptimeSeconds uint64, thresholdsInDays dynamotd.Thresholds) color.Attribute {
	const secondsInADay = 24 * 60 * 60
	return thresholdsInDays.Color(float64(uptimeSeconds) / secondsInADay)
}

func GetUptimeString(uptimeSeconds uint64) string {
	if uptimeSeconds < 60 {
		return "less than 1 minute"
	}

	uptime := getUptimeValues(uptimeSeconds)

	units := []struct {
		name  string
		value uint64
	}{
		{"year", uptime.years},
		{"month", uptime.months},
		{"day", uptime.days},
		{"hour", uptime.hours},
		{"minute", uptime.minutes},
	}

	var parts []string
	for _, unit := range units {
		switch unit.value {
		case 0:
			continue
		case 1:
			parts = append(parts, fmt.Sprintf("%d %s", unit.value, unit.name))
		default:
			parts = append(parts, fmt.Sprintf("%d %ss", unit.value, unit.name))
		}
	}

	return strings.Join(parts, ", ")
}

type uptimeValues struct {
	years, months, days, hours, minutes, seconds uint64
}

func getUptimeValues(uptimeSeconds uint64) uptimeValues {
	const (
		secondsInAMinute = 60
		secondsInAnHour  = 60 * secondsInAMinute
		secondsInADay    = 24 * secondsInAnHour
		secondsInAMonth  = 30 * secondsInADay
		secondsInAYear   = 365 * secondsInADay
	)

	var values uptimeValues
	remaining := uptimeSeconds

	values.years, remaining = remaining/secondsInAYear, remaining%secondsInAYear
	values.months, remaining = remaining/secondsInAMonth, remaining%secondsInAMonth
	values.days, remaining = remaining/secondsInADay, remaining%secondsInADay
	values.hours, remaining = remaining/secondsInAnHour, remaining%secondsInAnHour
	values.minutes, values.seconds = remaining/secondsInAMinute, remaining%secondsInAMinute

	return values
}
