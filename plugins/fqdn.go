package plugins

import (
	. "github.com/gdubicki/dynamotd/dynamotd"
	"fmt"
	gofqdn "github.com/Showmax/go-fqdn"
)

func Fqdn() Row {
	fqdnValue := gofqdn.Get()
	return Row {
		Label: SingleColorLabel("FQDN"),
		Value: SingleColorValue(fmt.Sprintf("%s", fqdnValue)),
	}
}
