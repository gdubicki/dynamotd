package main


import (
	"fmt"
	"github.com/spf13/viper"
)

var MaxRows = 40

var config = viper.New()

func configure() {
	config.SetDefault("mode", "static")
	config.SetDefault("rows", []string{
		"timestamp",
		"emptyLine",
		"fqdn",
		"emptyLine",
		"load",
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

func isModeStatic() bool {
	mode := config.GetString("mode")
	return mode == "static"
}

func getRows() []Row {
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
		case "emptyLine":
			rows = append(rows, emptyLine())
		case "timestamp":
			rows = append(rows, timestamp())
		case "fqdn":
			rows = append(rows, fqdn())
		case "load":
			rows = append(rows, load())
		default:
			panic(fmt.Errorf("error while generating row from string '%s' - check for typos", rowString))
		}
	}

	return rows
}
