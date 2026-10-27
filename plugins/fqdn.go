package plugins

import (
	gofqdn "github.com/Showmax/go-fqdn"
	"github.com/gdubicki/dynamotd/dynamotd"
)

func init() {
	dynamotd.Register(dynamotd.RowType{Label: fqdnLabel, Generate: dynamotd.SingleRow(Fqdn)}, "fqdn")
}

const fqdnLabel = "FQDN"

func Fqdn() dynamotd.Row {
	return dynamotd.Row{
		Label: dynamotd.SingleColorLabel(fqdnLabel),
		Value: dynamotd.SingleColorValue(gofqdn.Get()),
	}
}
