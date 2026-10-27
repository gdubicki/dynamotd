package plugins

import (
	"time"

	"github.com/gdubicki/dynamotd/dynamotd"
)

func init() {
	dynamotd.Register(Timestamp, "timestamp")
}

func Timestamp() dynamotd.Row {
	timestamp := time.Now().Format("2006-01-02 15:04:05 MST")
	return dynamotd.Row{
		Label: dynamotd.SingleColorLabel("Information as of"),
		Value: dynamotd.SingleColorValue(timestamp),
	}
}
