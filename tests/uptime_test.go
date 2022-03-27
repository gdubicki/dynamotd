package plugins

import (
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

func TestUptimeHourSingular(t *testing.T) {
	uptimeSeconds := uint64(4601)
	expectedString := "1 hour, 16 minutes"

	actualString := GetUptimeString(uptimeSeconds)

	assert.Equal(t, expectedString, actualString, "The two strings should be the same.")
}

func TestUptimeHours(t *testing.T) {
	uptimeSeconds := uint64(12000882)
	expectedString := "4 months, 18 days, 21 hours, 34 minutes"

	actualString := GetUptimeString(uptimeSeconds)

	assert.Equal(t, expectedString, actualString, "The two strings should be the same.")
}
