package dynamotd

import (
	"context"
	"testing"

	"github.com/fatih/color"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestThresholdsColor(t *testing.T) {
	thresholds := Thresholds{Warning: 75, Critical: 90}
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
		assert.Equal(t, tt.expected, thresholds.Color(tt.value), "value %v", tt.value)
	}
}

func settingsNode(t *testing.T, settingsYAML string) *yaml.Node {
	var node yaml.Node
	require.NoError(t, yaml.Unmarshal([]byte(settingsYAML), &node))
	// unwrap the document node
	return node.Content[0]
}

func TestDecodeSettings(t *testing.T) {
	defaults := Thresholds{Warning: 75, Critical: 90}

	settings, err := decodeSettings(defaults, nil)
	require.NoError(t, err)
	assert.Equal(t, defaults, settings)

	// the settings not set keep their default values
	settings, err = decodeSettings(defaults, settingsNode(t, "warning: 50"))
	require.NoError(t, err)
	assert.Equal(t, Thresholds{Warning: 50, Critical: 90}, settings)

	_, err = decodeSettings(defaults, settingsNode(t, "warnign: 50"))
	assert.ErrorContains(t, err, "field warnign not found")

	_, err = decodeSettings(defaults, settingsNode(t, "warning: 95"))
	assert.ErrorContains(t, err, "warning threshold (95) is higher than the critical one (90)")

	_, err = decodeSettings(NoSettings{}, settingsNode(t, "foo: bar"))
	assert.ErrorContains(t, err, "field foo not found")
}

func TestLookupWithSettings(t *testing.T) {
	Register("Test", Thresholds{Warning: 1, Critical: 2}, func(_ context.Context, thresholds Thresholds) []Row {
		return []Row{{Label: SingleColorLabel("Test"), Value: ToColorText(ColorString{Color: thresholds.Color(1.5), Text: "1.5"})}}
	}, "test-settings")

	rowTypes, err := Lookup([]RowConfig{{Name: "test-settings", Settings: settingsNode(t, "critical: 1.5")}})
	require.NoError(t, err)
	rows := rowTypes[0].Generate(context.Background())
	assert.Equal(t, ValueCriticalColor, rows[0].Value.text[0].Color)

	_, err = Lookup([]RowConfig{{Name: "test-settings", Settings: settingsNode(t, "critical: 0.5")}})
	assert.ErrorContains(t, err, "invalid settings for row type 'test-settings': warning threshold")
}
