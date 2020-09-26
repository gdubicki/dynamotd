package plugins

import (
	. "github.com/gdubicki/dynamotd/dynamotd"
)

func EmptyLine() Row {
	return Row {
		Label: SingleColorLabel(""),
		Value: SingleColorValue(""),
	}
}