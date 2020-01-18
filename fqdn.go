package main

import (
	"fmt"
	gofqdn "github.com/Showmax/go-fqdn"
)

func fqdn() Row {
	fqdnValue := gofqdn.Get()
	return Row{
		singleColorLabel("FQDN"),
		singleColorValue(fmt.Sprintf("%s", fqdnValue)),
	}
}
