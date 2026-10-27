package plugins

import (
	"github.com/gdubicki/dynamotd/dynamotd"
)

func init() {
	dynamotd.Register("", dynamotd.NoSettings{}, dynamotd.SingleRow(EmptyLine), "", "emptyLine")
}

func EmptyLine(dynamotd.NoSettings) dynamotd.Row {
	return dynamotd.Row{}
}
