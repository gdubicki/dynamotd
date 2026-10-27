package plugins

import (
	"github.com/gdubicki/dynamotd/dynamotd"
)

func init() {
	dynamotd.Register(EmptyLine, "", "emptyLine")
}

func EmptyLine() dynamotd.Row {
	return dynamotd.Row{}
}
