package plugins

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/gdubicki/dynamotd/dynamotd"
)

func init() {
	dynamotd.Register(Ip, "ip")
}

func Ip() dynamotd.Row {
	return dynamotd.Row{
		Label: dynamotd.SingleColorLabel("IP"),
		Value: dynamotd.ToColorText(
			ipColorString(getLocalIP()),
			dynamotd.ValueDescription(" (external: "),
			ipColorString(getExternalIP()),
			dynamotd.ValueDescription(")"),
		),
	}
}

func ipColorString(ip string, err error) dynamotd.ColorString {
	if err != nil {
		return dynamotd.ValueCritical(fmt.Sprintf("can't get: %s", err))
	}
	return dynamotd.ValueNeutral(ip)
}

// get preferred outbound ip of this machine
// copied from https://stackoverflow.com/a/37382208/2693875
func getLocalIP() (string, error) {
	// this does not really make the connection
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return "", err
	}
	defer func() { _ = conn.Close() }()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP.String(), nil
}

var showFakeIP = false

func getExternalIP() (string, error) {
	if showFakeIP {
		return "123.123.123.123", nil
	}

	client := http.Client{
		Timeout: 3 * time.Second,
	}
	resp, err := client.Get("https://icanhazip.com")
	if err != nil {
		return "", fmt.Errorf("no or too slow internet connection")
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected HTTP status %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("check internet connection")
	}
	return strings.TrimSpace(string(body)), nil
}
