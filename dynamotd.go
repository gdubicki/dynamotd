package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/gdubicki/dynamotd/dynamotd"
	_ "github.com/gdubicki/dynamotd/plugins"
)

// set during the build
var version = "dev"

func main() {
	var flagForceColor = flag.Bool("force-color", false, "Force color output")
	var flagNoColor = flag.Bool("no-color", false, "Disable color output")
	var flagVersion = flag.Bool("version", false, "Print version and exit")

	flag.Parse()

	if *flagVersion {
		fmt.Println(version)
		return
	}

	if *flagNoColor {
		color.NoColor = true
	} else if *flagForceColor {
		color.NoColor = false
	}

	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "dynamotd: %s\n", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	rowTypes, err := dynamotd.Lookup(cfg.Rows)
	if err != nil {
		return err
	}

	dynamotd.Print(dynamotd.Generate(rowTypes, cfg.Timeout))
	return nil
}
