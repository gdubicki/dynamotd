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

	"github.com/gdubicki/dynamotd/dynamotd"
)

func init() {
	dynamotd.Register(ipLabel, dynamotd.NoSettings{}, func(ctx context.Context, _ dynamotd.NoSettings) []dynamotd.Row {
		return []dynamotd.Row{Ip(ctx)}
	}, "ip")
}

const ipLabel = "IP"

func Ip(ctx context.Context) dynamotd.Row {
	return dynamotd.Row{
		Label: dynamotd.SingleColorLabel(ipLabel),
		Value: dynamotd.ToColorText(
			ipColorString(getLocalIP()),
			dynamotd.ValueDescription(" (external: "),
			ipColorString(getExternalIP(ctx)),
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
		return "", errors.New("no default route")
	}
	defer func() { _ = conn.Close() }()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP.String(), nil
}

var showFakeIP = false

func getExternalIP(ctx context.Context) (string, error) {
	if showFakeIP {
		return "123.123.123.123", nil
	}

	// the overall timeout is in the context
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://icanhazip.com", nil)
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
	return strings.TrimSpace(string(body)), nil
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
