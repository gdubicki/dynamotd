package main

import (
	"fmt"
	. "github.com/gdubicki/dynamotd/dynamotd"
	"runtime"
)

func init() {
	if runtime.GOOS != "linux" && runtime.GOOS != "darwin" {
		panic(fmt.Errorf("only Linux and macOS are supported by this app"))
	}
}

func main() {
	rows := GetRows()

	if IsModeStatic() {
		PrintStatic(rows)
	} else {
		ShowDynamic(rows)
	}
}
