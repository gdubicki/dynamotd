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
		localIPcolor = valueCriticalColor
	} else {
		localIPcolor = valueNeutralColor
	}

	if strings.Contains(externalIP, "can't") {
		externalIPcolor = valueCriticalColor
	} else {
		externalIPcolor = valueNeutralColor
	}

	return Row{
		singleColorLabel("IP"),
		toColorText(
			ColorString{localIPcolor, localIP},
			valueDescription(" (external: "),
			ColorString{externalIPcolor, externalIP},
			valueDescription(")"),
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