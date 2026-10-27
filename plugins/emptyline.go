package plugins

import (
	"github.com/gdubicki/dynamotd/dynamotd"
)

func init() {
	dynamotd.Register(dynamotd.RowType{Label: "", Generate: dynamotd.SingleRow(EmptyLine)}, "", "emptyLine")
}

func EmptyLine() dynamotd.Row {
	return dynamotd.Row{}
}
