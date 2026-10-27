package main

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseConfig(t *testing.T) {
	cfg, err := parseConfig([]byte("rows:\n  - timestamp\n  -\n  - fqdn\n"), "test.yaml")
	require.NoError(t, err)
	assert.Equal(t, []string{"timestamp", "", "fqdn"}, cfg.Rows)
}

func TestParseConfigWithoutRowsUsesDefaults(t *testing.T) {
	cfg, err := parseConfig([]byte("# nothing here\n"), "test.yaml")
	require.NoError(t, err)
	assert.Equal(t, defaultRows, cfg.Rows)
}

func TestParseConfigErrors(t *testing.T) {
	_, err := parseConfig([]byte("rows: []\n"), "test.yaml")
	assert.ErrorContains(t, err, "no rows defined")

	_, err = parseConfig([]byte("rows: [\n"), "test.yaml")
	assert.ErrorContains(t, err, "error in config file test.yaml")
}

// the example config in the repo should be the same as the defaults, as it says
func TestExampleConfigMatchesDefaults(t *testing.T) {
	data, err := os.ReadFile(configFileName)
	require.NoError(t, err)
	cfg, err := parseConfig(data, configFileName)
	require.NoError(t, err)
	assert.Equal(t, defaultConfig(), cfg)
}

func TestParseConfigTimeout(t *testing.T) {
	cfg, err := parseConfig([]byte("timeout: 2500ms\n"), "test.yaml")
	require.NoError(t, err)
	assert.Equal(t, 2500*time.Millisecond, cfg.Timeout)
	assert.Equal(t, defaultRows, cfg.Rows)

	cfg, err = parseConfig([]byte("rows: [timestamp]\n"), "test.yaml")
	require.NoError(t, err)
	assert.Equal(t, defaultTimeout, cfg.Timeout)

	_, err = parseConfig([]byte("timeout: 0s\n"), "test.yaml")
	assert.ErrorContains(t, err, "timeout must be positive")
}

func TestParseConfigUnknownKey(t *testing.T) {
	_, err := parseConfig([]byte("timout: 2s\n"), "test.yaml")
	assert.ErrorContains(t, err, "field timout not found")
}

func TestParseConfigEmptyFile(t *testing.T) {
	cfg, err := parseConfig([]byte(""), "test.yaml")
	require.NoError(t, err)
	assert.Equal(t, defaultConfig(), cfg)
}
