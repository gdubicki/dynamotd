package plugins

import (
	"time"

	"github.com/gdubicki/dynamotd/dynamotd"
)

func init() {
	dynamotd.Register(dynamotd.RowType{Label: timestampLabel, Generate: dynamotd.SingleRow(Timestamp)}, "timestamp")
}

const timestampLabel = "Information as of"

func Timestamp() dynamotd.Row {
	timestamp := time.Now().Format("2006-01-02 15:04:05 MST")
	return dynamotd.Row{
		Label: dynamotd.SingleColorLabel(timestampLabel),
		Value: dynamotd.SingleColorValue(timestamp),
	}
}
