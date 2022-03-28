package plugins

import (
	"fmt"
	. "github.com/gdubicki/dynamotd/dynamotd"
	"time"
)

func Timestamp() Row {
	timestamp := time.Now().Format("2006-01-02 15:04:05 MST")
	return Row{
		Label: SingleColorLabel("Information as of"),
		Value: SingleColorValue(fmt.Sprintf("%s", timestamp)),
	}
}
