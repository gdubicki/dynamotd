package main

import (
	"fmt"
	gofqdn "github.com/Showmax/go-fqdn"
)

func fqdn() Row {
	fqdnValue := gofqdn.Get()
	return Row{
		singleColorLabelText("FQDN"),
		singleColorValueText(fmt.Sprintf("%s", fqdnValue)),
	}
}
