package main

import (
	"fmt"
	"time"
)

func timestamp() Row {
	timestamp := time.Now().Format("2006-01-02 15:04:05 -07:00 MST")
	return Row{
		singleColorLabel("Information as of"),
		singleColorValue(fmt.Sprintf("%s", timestamp)),
	}
}
