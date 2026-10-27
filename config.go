package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"syscall"
	"time"

	"github.com/gdubicki/dynamotd/dynamotd"
	"github.com/gdubicki/dynamotd/plugins"
	"gopkg.in/yaml.v3"
)

const maxRows = 40

const configFileName = "dynamotd.yaml"

var defaultRows = rowConfigs(
	"timestamp",
	"",
	"fqdn",
	"ip",
	"uptime",
	"users",
	"reboot",
	"clock",
	"",
	"load",
	"cpu",
	"memory",
	"topprocesses",
	"diskspace",
	"failedunits",
)

func rowConfigs(names ...string) []dynamotd.RowConfig {
	rows := make([]dynamotd.RowConfig, 0, len(names))
	for _, name := range names {
		rows = append(rows, dynamotd.RowConfig{Name: name})
	}
	return rows
}

// the rows that don't finish in this time are shown as timed out, so that a login is never blocked for long
const defaultTimeout = 1 * time.Second

type config struct {
	// Path is the config file used, empty if none
	Path    string
	Rows    []dynamotd.RowConfig
	Timeout time.Duration
}

func defaultConfig() config {
	return config{Rows: defaultRows, Timeout: defaultTimeout}
}

// configPaths returns the locations of the config file, in the order they are searched.
// The current directory is deliberately not one of them, as this app usually runs as root.
func configPaths() []string {
	paths := []string{filepath.Join("/etc", configFileName)}
	if home, err := os.UserHomeDir(); err == nil {
		paths = append(paths, filepath.Join(home, ".dynamotd", configFileName))
	}
	return paths
}

// loadConfig reads the config file from the given path, or if it's empty - the first config file found
// in the default locations. The config file is optional - the defaults are enough to run this app.
func loadConfig(path string) (config, error) {
	if path != "" {
		return readConfigFile(path)
	}

	for _, path := range configPaths() {
		cfg, err := readConfigFile(path)
		if errors.Is(err, fs.ErrNotExist) {
			continue
		}
		return cfg, err
	}
	return defaultConfig(), nil
}

func readConfigFile(path string) (config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return config{}, err
	}
	cfg, err := parseConfig(data, path)
	if err != nil {
		return config{}, err
	}

	for _, row := range cfg.Rows {
		if row.Name == plugins.CommandRowType {
			if err := checkConfigFileIsSecure(path); err != nil {
				return config{}, err
			}
			break
		}
	}
	return cfg, nil
}

// checkConfigFileIsSecure checks if the config file can only be changed by root or the current user,
// as the commands in it are run as the user running this app, which is usually root.
func checkConfigFileIsSecure(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	stat, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		return fmt.Errorf("can't check the owner of the config file %s", path)
	}
	if stat.Uid != 0 && int(stat.Uid) != os.Geteuid() {
		return fmt.Errorf("config file %s has %s rows, so it must be owned by root or the current user", path, plugins.CommandRowType)
	}
	if info.Mode().Perm()&0o022 != 0 {
		return fmt.Errorf("config file %s has %s rows, so it must not be writable by group or others", path, plugins.CommandRowType)
	}
	return nil
}

func parseConfig(data []byte, path string) (config, error) {
	var raw struct {
		// nodes, as the rows can be strings, nulls (empty lines) or maps (with settings)
		Rows    *[]yaml.Node   `yaml:"rows"`
		Timeout *time.Duration `yaml:"timeout"`
	}
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	// report typos in the keys
	decoder.KnownFields(true)
	if err := decoder.Decode(&raw); err != nil && !errors.Is(err, io.EOF) {
		return config{}, fmt.Errorf("error in config file %s: %w", path, err)
	}

	cfg := defaultConfig()
	cfg.Path = path

	if raw.Timeout != nil {
		if *raw.Timeout <= 0 {
			return config{}, fmt.Errorf("error in config file %s: timeout must be positive", path)
		}
		cfg.Timeout = *raw.Timeout
	}

	// nil means no "rows" key at all, so use the defaults
	if raw.Rows == nil {
		return cfg, nil
	}

	rows := make([]dynamotd.RowConfig, 0, len(*raw.Rows))
	for _, node := range *raw.Rows {
		row, err := parseRow(&node)
		if err != nil {
			return config{}, fmt.Errorf("error in config file %s, line %d: %w", path, node.Line, err)
		}
		rows = append(rows, row)
	}

	if len(rows) == 0 {
		return config{}, fmt.Errorf("error in config file %s: no rows defined", path)
	}
	if len(rows) > maxRows {
		return config{}, fmt.Errorf("error in config file %s: more rows than %d not supported", path, maxRows)
	}
	cfg.Rows = rows
	return cfg, nil
}

// parseRow parses a row, which can be:
// * a row type name,
// * empty, for an empty line,
// * a map with a single key - the row type name, and its settings as the value.
func parseRow(node *yaml.Node) (dynamotd.RowConfig, error) {
	switch {
	case node.Kind == yaml.ScalarNode && node.Tag == "!!null":
		return dynamotd.RowConfig{Name: ""}, nil
	case node.Kind == yaml.ScalarNode:
		return dynamotd.RowConfig{Name: node.Value}, nil
	case node.Kind == yaml.MappingNode && len(node.Content) == 2:
		name, settings := node.Content[0], node.Content[1]
		if settings.Kind == yaml.ScalarNode && settings.Tag == "!!null" {
			settings = nil
		}
		return dynamotd.RowConfig{Name: name.Value, Settings: settings}, nil
	default:
		return dynamotd.RowConfig{}, fmt.Errorf("a row should be a row type name or a map with a row type name as the only key and its settings as the value")
	}
}
