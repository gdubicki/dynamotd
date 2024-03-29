package main

import (
	"fmt"
	. "github.com/gdubicki/dynamotd/dynamotd"
	. "github.com/gdubicki/dynamotd/plugins"
	"github.com/spf13/viper"
)

var MaxRows = 40

var config = viper.New()

func init() {
	config.SetDefault("rows", []string{
		"timestamp",
		"",
		"fqdn",
		"ip",
		"uptime",
		"",
		"load",
		"memory",
		"diskspace",
	})

	config.SetConfigName("dynamotd.yaml")
	config.AddConfigPath("/etc/")
	config.AddConfigPath("$HOME/.dynamotd")
	config.AddConfigPath(".")

	err := config.ReadInConfig()
	if err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// that's ok - we don't need a config file for this app. the defaults are enough to run it.
		} else {
			panic(fmt.Errorf("Fatal error in config file: %s \n", err))
		}
	}
}

func GetRows() []Row {
	var rows []Row

	rowsStrings := config.GetStringSlice("rows")

	if len(rowsStrings) == 0 {
		panic(fmt.Errorf("no rows defined"))
	}

	if len(rowsStrings) > MaxRows {
		panic(fmt.Errorf("more rows than %d not supported", MaxRows))
	}

	for _, rowString := range rowsStrings {

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
