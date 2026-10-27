package main

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	. "github.com/gdubicki/dynamotd/dynamotd"
	. "github.com/gdubicki/dynamotd/plugins"
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

type config struct {
	Rows []string `yaml:"rows"`
}

// configPaths returns the locations of the config file, in the order they are searched.
func configPaths() []string {
	paths := []string{filepath.Join("/etc", configFileName)}
	if home, err := os.UserHomeDir(); err == nil {
		paths = append(paths, filepath.Join(home, ".dynamotd", configFileName))
	}
	return append(paths, configFileName)
}

// loadConfig reads the first config file found. The config file is optional -
// the defaults are enough to run this app.
func loadConfig() (config, error) {
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
	return config{Rows: defaultRows}, nil
}

func parseConfig(data []byte, path string) (config, error) {
	var raw struct {
		// pointers, as otherwise the empty (null) elements, used for empty lines, would be skipped
		Rows *[]*string `yaml:"rows"`
	}
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return config{}, fmt.Errorf("error in config file %s: %w", path, err)
	}

	// nil means no "rows" key at all, so use the defaults
	if raw.Rows == nil {
		return config{Rows: defaultRows}, nil
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
	return config{Rows: rows}, nil
}

func GetRows() []Row {
	var rows []Row

	cfg, err := loadConfig()
	if err != nil {
		panic(err)
	}

	for _, rowString := range cfg.Rows {

		// TODO: consider using reflection to be able to add row types without editing this file
		switch rowString {
		case "", "emptyLine":
			rows = append(rows, EmptyLine())
		case "timestamp":
			rows = append(rows, Timestamp())
		case "fqdn":
			rows = append(rows, Fqdn())
		case "load":
			rows = append(rows, Load())
		case "ip":
			rows = append(rows, Ip())
		case "uptime":
			rows = append(rows, Uptime())
		case "cores":
			rows = append(rows, Cores())
		case "memory":
			rows = append(rows, Memory())
		case "diskspace":
			rows = append(rows, DiskSpace("/"))
		default:
			panic(fmt.Errorf("error while generating row from string '%s' - check for typos", rowString))
		}
	}

	return rows
}
