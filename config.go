package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

const maxRows = 40

const configFileName = "dynamotd.yaml"

var defaultRows = []string{
	"timestamp",
	"",
	"fqdn",
	"ip",
	"uptime",
	"",
	"load",
	"memory",
	"diskspace",
}

// the rows that don't finish in this time are shown as timed out, so that a login is never blocked for long
const defaultTimeout = 1 * time.Second

type config struct {
	Rows    []string
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
		data, err := os.ReadFile(path)
		if err != nil {
			return config{}, err
		}
		return parseConfig(data, path)
	}

	for _, path := range configPaths() {
		data, err := os.ReadFile(path)
		if errors.Is(err, fs.ErrNotExist) {
			continue
		}
		if err != nil {
			return config{}, err
		}
		return parseConfig(data, path)
	}
	return defaultConfig(), nil
}

func parseConfig(data []byte, path string) (config, error) {
	var raw struct {
		// pointers, as otherwise the empty (null) elements, used for empty lines, would be skipped
		Rows    *[]*string     `yaml:"rows"`
		Timeout *time.Duration `yaml:"timeout"`
	}
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	// report typos in the keys
	decoder.KnownFields(true)
	if err := decoder.Decode(&raw); err != nil && !errors.Is(err, io.EOF) {
		return config{}, fmt.Errorf("error in config file %s: %w", path, err)
	}

	cfg := defaultConfig()

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

	rows := make([]string, 0, len(*raw.Rows))
	for _, row := range *raw.Rows {
		if row == nil {
			rows = append(rows, "")
		} else {
			rows = append(rows, *row)
		}
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
