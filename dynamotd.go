package main

import (
	. "github.com/gdubicki/dynamotd/dynamotd"
)

func main() {
	rows := GetRows()

	if IsModeStatic() {
		PrintStatic(rows)
	} else {
		ShowDynamic(rows)
	}
}
