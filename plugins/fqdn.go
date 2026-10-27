package plugins

import (
	gofqdn "github.com/Showmax/go-fqdn"
	"github.com/gdubicki/dynamotd/dynamotd"
)

func init() {
	dynamotd.Register(Fqdn, "fqdn")
}

func Fqdn() dynamotd.Row {
	return dynamotd.Row{
		Label: dynamotd.SingleColorLabel("FQDN"),
		Value: dynamotd.SingleColorValue(gofqdn.Get()),
	}
}
