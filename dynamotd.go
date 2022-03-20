package main

import (
	"flag"
	"fmt"
	"github.com/fatih/color"
	. "github.com/gdubicki/dynamotd/dynamotd"
	"runtime"
)

func init() {
	if runtime.GOOS != "linux" && runtime.GOOS != "darwin" {
		panic(fmt.Errorf("only Linux and macOS are supported by this app"))
	}
}

func main() {
	var flagForceColor = flag.Bool("force-color", false, "Force color output")
	var flagNoColor = flag.Bool("no-color", false, "Disable color output")

	flag.Parse()

	if *flagNoColor {
		color.NoColor = true
	} else if *flagForceColor {
		color.NoColor = false
	}

	rows := GetRows()
	if IsModeStatic() {
		PrintStatic(rows)
	} else {
		ShowDynamic(rows)
	}
}
