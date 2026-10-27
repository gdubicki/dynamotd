package main

import (
	"io/fs"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/gdubicki/dynamotd/dynamotd"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestParseConfig(t *testing.T) {
	cfg, err := parseConfig([]byte("rows:\n  - timestamp\n  -\n  - fqdn\n"), "test.yaml")
	require.NoError(t, err)
	assert.Equal(t, rowConfigs("timestamp", "", "fqdn"), cfg.Rows)
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
	cfg.Path = ""
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
	assert.Equal(t, "test.yaml", cfg.Path)
	cfg.Path = ""
	assert.Equal(t, defaultConfig(), cfg)
}

func TestConfigPathsDontIncludeCurrentDirectory(t *testing.T) {
	for _, path := range configPaths() {
		assert.True(t, filepath.IsAbs(path), path)
	}
}

func TestLoadConfigFromPath(t *testing.T) {
	path := filepath.Join(t.TempDir(), "custom.yaml")
	require.NoError(t, os.WriteFile(path, []byte("rows: [fqdn]\n"), 0o644))

	cfg, err := loadConfig(path)
	require.NoError(t, err)
	assert.Equal(t, rowConfigs("fqdn"), cfg.Rows)
	assert.Equal(t, path, cfg.Path)

	_, err = loadConfig(filepath.Join(t.TempDir(), "missing.yaml"))
	assert.ErrorIs(t, err, fs.ErrNotExist)
}

func TestParseConfigRowsWithSettings(t *testing.T) {
	cfg, err := parseConfig([]byte(`
rows:
  - memory:
      warning: 80
  - load:
  - fqdn
`), "test.yaml")
	require.NoError(t, err)
	require.Len(t, cfg.Rows, 3)

	assert.Equal(t, "memory", cfg.Rows[0].Name)
	require.NotNil(t, cfg.Rows[0].Settings)
	assert.Equal(t, yaml.MappingNode, cfg.Rows[0].Settings.Kind)

	// a row type name with an empty value is the same as just the name
	assert.Equal(t, dynamotd.RowConfig{Name: "load"}, cfg.Rows[1])
	assert.Equal(t, dynamotd.RowConfig{Name: "fqdn"}, cfg.Rows[2])
}

func TestParseConfigInvalidRow(t *testing.T) {
	_, err := parseConfig([]byte("rows:\n  - timestamp\n  - memory: {}\n    load: {}\n"), "test.yaml")
	assert.ErrorContains(t, err, "error in config file test.yaml, line 3: a row should be")
}

func TestCommandRowsRequireSecureConfigFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "dynamotd.yaml")
	require.NoError(t, os.WriteFile(path, []byte("rows:\n  - command:\n      label: X\n      run: id\n"), 0o644))

	_, err := loadConfig(path)
	require.NoError(t, err)

	require.NoError(t, os.Chmod(path, 0o664))
	_, err = loadConfig(path)
	assert.ErrorContains(t, err, "has command rows, so it must not be writable by group or others")

	require.NoError(t, os.Chmod(path, 0o646))
	_, err = loadConfig(path)
	assert.ErrorContains(t, err, "must not be writable by group or others")
}

func TestOtherRowsDontRequireSecureConfigFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "dynamotd.yaml")
	require.NoError(t, os.WriteFile(path, []byte("rows: [fqdn]\n"), 0o666))
	require.NoError(t, os.Chmod(path, 0o666))

	_, err := loadConfig(path)
	assert.NoError(t, err)
}
