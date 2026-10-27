package dynamotd

import (
	"errors"
	"testing"

	"github.com/fatih/color"
	"github.com/stretchr/testify/assert"
)

func TestThresholdColor(t *testing.T) {
	tests := []struct {
		value    float64
		expected color.Attribute
	}{
		{0, ValueOkColor},
		{74.9, ValueOkColor},
		{75, ValueWarningColor},
		{89.9, ValueWarningColor},
		{90, ValueCriticalColor},
		{100, ValueCriticalColor},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.expected, ThresholdColor(tt.value, 75, 90), "value %v", tt.value)
	}
}

func TestErrorRow(t *testing.T) {
	row := ErrorRow("RAM", errors.New("boom"))
	assert.Equal(t, "RAM", row.Label.String())
	assert.Equal(t, "can't get: boom", row.Value.String())
	assert.False(t, row.isEmptyLine())
}

func TestGetStringToPrint(t *testing.T) {
	color.NoColor = false
	defer func() { color.NoColor = true }()

	// percent signs must not be interpreted as format verbs
	assert.Equal(t, "100%", getStringToPrint(ValueDescription("100%")))
	assert.Equal(t, "\x1b[31m100%\x1b[0m", getStringToPrint(ValueCritical("100%")))
}
