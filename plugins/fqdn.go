package plugins

import (
	"os"

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
		Value: dynamotd.SingleColorValue(fqdn()),
	}
}

// fqdn matches `hostname -f`, unlike gofqdn.Get(), which returns the reverse lookup of the first IPv4 address.
func fqdn() string {
	if name, err := gofqdn.FqdnHostname(); err == nil {
		return name
	}
	if name, err := os.Hostname(); err == nil {
		return name
	}
	return "unknown"
}
