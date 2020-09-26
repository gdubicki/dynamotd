package plugins

import (
	"fmt"
	gofqdn "github.com/Showmax/go-fqdn"
	. "github.com/gdubicki/dynamotd/dynamotd"
)

func Fqdn() Row {
	fqdnValue := gofqdn.Get()
	return Row{
		Label: SingleColorLabel("FQDN"),
		Value: SingleColorValue(fmt.Sprintf("%s", fqdnValue)),
	}
}
