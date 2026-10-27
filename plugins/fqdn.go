package plugins

import (
	gofqdn "github.com/Showmax/go-fqdn"
	"github.com/gdubicki/dynamotd/dynamotd"
)

func init() {
	dynamotd.Register(fqdnLabel, dynamotd.NoSettings{}, dynamotd.SingleRow(Fqdn), "fqdn")
}

const fqdnLabel = "FQDN"

func Fqdn(dynamotd.NoSettings) dynamotd.Row {
	return dynamotd.Row{
		Label: dynamotd.SingleColorLabel(fqdnLabel),
		Value: dynamotd.SingleColorValue(gofqdn.Get()),
	}
}
