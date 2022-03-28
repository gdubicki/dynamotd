package plugins

import (
	. "github.com/gdubicki/dynamotd/dynamotd"
	. "github.com/gdubicki/dynamotd/plugins"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestUptimeSeconds(t *testing.T) {
	uptimeSeconds := uint64(15)
	expectedString := "less than 1 minute"

	actualString := GetUptimeString(uptimeSeconds)

	assert.Equal(t, expectedString, actualString, "The two strings should be the same.")
}

var overHourInSeconds = uint64(4601)

func TestUptimeHourSingular(t *testing.T) {
	expectedString := "1 hour, 16 minutes"

	actualString := GetUptimeString(overHourInSeconds)

	assert.Equal(t, expectedString, actualString, "The two strings should be the same.")
}

func TestUptimeHourColor(t *testing.T) {
	expectedColor := ValueOkColor

	actualColor := GetUptimeColor(overHourInSeconds)

	assert.Equal(t, expectedColor, actualColor, "The two colors should be the same.")
}

var overFourMonthsInSeconds = uint64(12000882)

func TestUptimeMonthsString(t *testing.T) {
	expectedString := "4 months, 18 days, 21 hours, 34 minutes"

	actualString := GetUptimeString(overFourMonthsInSeconds)

	assert.Equal(t, expectedString, actualString, "The two strings should be the same.")
}

func TestUptimeMonthsColor(t *testing.T) {
	expectedColor := ValueWarningColor

	actualColor := GetUptimeColor(overFourMonthsInSeconds)

	assert.Equal(t, expectedColor, actualColor, "The two colors should be the same.")
}
