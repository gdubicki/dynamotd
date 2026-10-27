package plugins

import (
	"testing"

	"github.com/fatih/color"
	"github.com/gdubicki/dynamotd/dynamotd"
	"github.com/stretchr/testify/assert"
)

func TestGetUptimeString(t *testing.T) {
	tests := []struct {
		name          string
		uptimeSeconds uint64
		expected      string
	}{
		{"seconds", 15, "less than 1 minute"},
		{"exactly a minute", 60, "1 minute"},
		{"hour singular", 4601, "1 hour, 16 minutes"},
		{"skips zero units", 2*24*60*60 + 5*60, "2 days, 5 minutes"},
		{"months", 12000882, "4 months, 18 days, 21 hours, 34 minutes"},
		{"year", 365*24*60*60 + 60*60, "1 year, 1 hour"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, GetUptimeString(tt.uptimeSeconds))
		})
	}
}

func TestGetUptimeColor(t *testing.T) {
	tests := []struct {
		name          string
		uptimeSeconds uint64
		expected      color.Attribute
	}{
		{"hour", 4601, dynamotd.ValueOkColor},
		{"just under 3 months", 3*30*24*60*60 - 1, dynamotd.ValueOkColor},
		{"over 4 months", 12000882, dynamotd.ValueWarningColor},
		{"over a year", 365*24*60*60 + 1, dynamotd.ValueCriticalColor},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, GetUptimeColor(tt.uptimeSeconds))
		})
	}
}

func TestGetUptimeValues(t *testing.T) {
	assert.Equal(t,
		uptimeValues{years: 1, months: 2, days: 3, hours: 4, minutes: 5, seconds: 6},
		getUptimeValues(365*24*60*60+2*30*24*60*60+3*24*60*60+4*60*60+5*60+6),
	)
}
