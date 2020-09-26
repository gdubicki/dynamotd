package plugins

import (
	. "github.com/gdubicki/dynamotd/dynamotd"
)

func EmptyLine() Row {
	return *new(Row)
}
