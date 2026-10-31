package plugins

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"sync"

	"github.com/gdubicki/dynamotd/dynamotd"
)

func init() {
	dynamotd.Register(ipLabel, dynamotd.NoSettings{}, func(ctx context.Context, _ dynamotd.NoSettings) []dynamotd.Row {
		return Ip(ctx)
	}, "ip")
}

const ipLabel = "IP"

type ipVersion struct {
	name string
	// any address outside of the local network, to find the local IP used to reach it
	remoteAddr string
	// returns the IP that the requests come from, without any other text
	externalIPURL string
}

var ipVersions = []ipVersion{
	{"IPv4", "8.8.8.8:80", "https://ipv4.icanhazip.com"},
	{"IPv6", "[2001:4860:4860::8888]:80", "https://ipv6.icanhazip.com"},
}

// for mocking in tests
var getLocalIP = getLocalIPReal
var getExternalIP = getExternalIPReal

// Ip shows the local and the external IPv4 and IPv6, or just one of them if they are the same,
// f.e. "Public IPv4: 203.0.113.45" or "IPv4: 10.0.1.23 (external: 203.0.113.45)".
// IPv6 is shown only if the machine has it.
func Ip(ctx context.Context) []dynamotd.Row {
	rows := make([]dynamotd.Row, len(ipVersions))
	shown := make([]bool, len(ipVersions))
	var wg sync.WaitGroup
	for i, version := range ipVersions {
		wg.Go(func() { rows[i], shown[i] = ipRow(ctx, version) })
	}
	wg.Wait()

	var result []dynamotd.Row
	for i, row := range rows {
		if shown[i] {
			result = append(result, row)
		}
	}
	return result
}

func ipRow(ctx context.Context, version ipVersion) (dynamotd.Row, bool) {
	row := func(label string, value ...dynamotd.ColorString) (dynamotd.Row, bool) {
		return dynamotd.Row{Label: dynamotd.SingleColorLabel(label), Value: dynamotd.ToColorText(value...)}, true
	}

	localIP, err := getLocalIP(version.remoteAddr)
	if err != nil {
		// many machines don't have IPv6, so it's not worth mentioning
		if version.name == "IPv6" {
			return dynamotd.Row{}, false
		}
		return row(version.name, ipColorString("", err))
	}

	externalIP, err := getExternalIP(ctx, version.externalIPURL)
	if err == nil && net.ParseIP(localIP).Equal(net.ParseIP(externalIP)) {
		return row("Public "+version.name, ipColorString(localIP, nil))
	}
	return row(version.name,
		ipColorString(localIP, nil),
		dynamotd.ValueDescription(" (external: "),
		ipColorString(externalIP, err),
		dynamotd.ValueDescription(")"),
	)
}

func ipColorString(ip string, err error) dynamotd.ColorString {
	if err != nil {
		return dynamotd.ValueCritical(fmt.Sprintf("can't get: %s", err))
	}
	return dynamotd.ValueNeutral(ip)
}

// get preferred outbound ip of this machine
// copied from https://stackoverflow.com/a/37382208/2693875
func getLocalIPReal(remoteAddr string) (string, error) {
	// this does not really make the connection
	conn, err := net.Dial("udp", remoteAddr)
	if err != nil {
		return "", errors.New("no default route")
	}
	defer func() { _ = conn.Close() }()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP.String(), nil
}

func getExternalIPReal(ctx context.Context, url string) (string, error) {
	// the overall timeout is in the context
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", describeHTTPError(err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected HTTP status %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("check internet connection")
	}
	ip := strings.TrimSpace(string(body))
	if net.ParseIP(ip) == nil {
		return "", fmt.Errorf("unexpected response")
	}
	return ip, nil
}

// describeHTTPError returns a short, human-readable reason for a failed HTTP request.
func describeHTTPError(err error) error {
	var dnsErr *net.DNSError
	var certErr *tls.CertificateVerificationError
	var opErr *net.OpError
	switch {
	case errors.Is(err, context.DeadlineExceeded):
		return fmt.Errorf("timed out")
	case errors.As(err, &dnsErr):
		return fmt.Errorf("can't resolve %s", dnsErr.Name)
	case errors.As(err, &certErr):
		return fmt.Errorf("TLS certificate verification failed - missing CA certificates?")
	case errors.As(err, &opErr):
		return fmt.Errorf("no internet connection")
	default:
		return err
	}
}
