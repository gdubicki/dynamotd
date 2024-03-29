package plugins

import (
	"github.com/fatih/color"
	. "github.com/gdubicki/dynamotd/dynamotd"
	"io"
	"net"
	"net/http"
	"strings"
	"time"
)

func Ip() Row {
	localIP := getLocalIP()
	externalIP := getExternalIP()

	var localIPcolor, externalIPcolor color.Attribute

	if strings.Contains(localIP, "can't") {
		localIPcolor = ValueCriticalColor
	} else {
		localIPcolor = ValueNeutralColor
	}

	if strings.Contains(externalIP, "can't") {
		externalIPcolor = ValueCriticalColor
	} else {
		externalIPcolor = ValueNeutralColor
	}

	return Row{
		Label: SingleColorLabel("IP"),
		Value: ToColorText(
			ColorString{Color: localIPcolor, Text: localIP},
			ValueDescription(" (external: "),
			ColorString{Color: externalIPcolor, Text: externalIP},
			ValueDescription(")"),
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

var showFakeIP = false

func getExternalIP() string {
	if showFakeIP {
		return "123.123.123.123"
	} else {
		client := http.Client{
			Timeout: 3 * time.Second,
		}
		resp, err := client.Get("https://icanhazip.com")
		if err != nil {
			return "can't get / too slow internet connection!"
		}
		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return "can't get - check internet connection!"
		}
		return strings.TrimSpace(string(body))
	}
}
