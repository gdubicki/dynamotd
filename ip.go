package main

import (
	"io/ioutil"
	"net"
	"net/http"
	"strings"
)


func ip() Row {
	localIP := getLocalIP()
	externalIP := getExternalIP()

	var localIPcolor, externalIPcolor Color

	if strings.Contains(localIP, "can't") {
		localIPcolor = keyValueCriticalColor
	} else {
		localIPcolor = keyValueOkColor
	}

	if strings.Contains(externalIP, "can't") {
		externalIPcolor = keyValueCriticalColor
	} else {
		externalIPcolor = keyValueOkColor
	}

	return Row{
		singleColorLabelText("IP"),
		toColorText(
			ColorString{localIPcolor, localIP},
			value(" (external: "),
			ColorString{externalIPcolor, externalIP},
			value(")"),
		),
	}
}

// get preferred outbound ip of this machine
// copied from https://stackoverflow.com/a/37382208/2693875
func getLocalIP() string {
	// this does not really make the connection

	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return "can't get"
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP.String()
}

func getExternalIP() string {
	resp, err := http.Get("https://icanhazip.com")
	if err != nil {
		return "can't get - check internet connection!"
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "can't get - check internet connection!"
	}
	return strings.TrimSpace(string(body))
}