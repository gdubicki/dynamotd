package main

import (
	"fmt"
	"time"
)

func timestamp() Row {
	timestamp := time.Now()
	return Row{
		singleColorLabelText("Information as of"),
		singleColorValueText(fmt.Sprintf("%s", timestamp)),
	}
}
