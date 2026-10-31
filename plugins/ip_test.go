package plugins

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/url"
	"testing"

	"github.com/gdubicki/dynamotd/dynamotd"
	"github.com/stretchr/testify/assert"
)

// mockIPs mocks the local and external IPs, by IP version - "" means "can't get it"
func mockIPs(t *testing.T, local, external map[string]string) {
	origLocal, origExternal := getLocalIP, getExternalIP
	t.Cleanup(func() { getLocalIP, getExternalIP = origLocal, origExternal })

	versionOf := func(target string) string {
		for _, version := range ipVersions {
			if target == version.remoteAddr || target == version.externalIPURL {
				return version.name
			}
		}
		t.Fatalf("unexpected target %s", target)
		return ""
	}
	getLocalIP = func(remoteAddr string) (string, error) {
		if ip := local[versionOf(remoteAddr)]; ip != "" {
			return ip, nil
		}
		return "", errors.New("no default route")
	}
	getExternalIP = func(_ context.Context, url string) (string, error) {
		if ip := external[versionOf(url)]; ip != "" {
			return ip, nil
		}
		return "", errors.New("timed out")
	}
}

func rowsText(rows []dynamotd.Row) []string {
	var texts []string
	for _, row := range rows {
		texts = append(texts, row.Label.String()+": "+row.Value.String())
	}
	return texts
}

func TestIp(t *testing.T) {
	tests := []struct {
		name            string
		local, external map[string]string
		expected        []string
	}{
		{
			name:     "behind NAT, without IPv6",
			local:    map[string]string{"IPv4": "10.0.1.23"},
			external: map[string]string{"IPv4": "203.0.113.45"},
			expected: []string{"IPv4: 10.0.1.23 (external: 203.0.113.45)"},
		},
		{
			name:     "public IPv4 and IPv6",
			local:    map[string]string{"IPv4": "203.0.113.45", "IPv6": "2001:db8::1"},
			external: map[string]string{"IPv4": "203.0.113.45", "IPv6": "2001:0db8:0000::1"},
			expected: []string{"Public IPv4: 203.0.113.45", "Public IPv6: 2001:db8::1"},
		},
		{
			name:     "IPv4 behind NAT, public IPv6",
			local:    map[string]string{"IPv4": "192.168.1.10", "IPv6": "2001:db8::1"},
			external: map[string]string{"IPv4": "203.0.113.45", "IPv6": "2001:db8::1"},
			expected: []string{"IPv4: 192.168.1.10 (external: 203.0.113.45)", "Public IPv6: 2001:db8::1"},
		},
		{
			name:     "no internet",
			local:    map[string]string{"IPv4": "10.0.1.23", "IPv6": "2001:db8::1"},
			expected: []string{"IPv4: 10.0.1.23 (external: can't get: timed out)", "IPv6: 2001:db8::1 (external: can't get: timed out)"},
		},
		{
			name:     "no network",
			expected: []string{"IPv4: can't get: no default route"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockIPs(t, tt.local, tt.external)
			assert.Equal(t, tt.expected, rowsText(Ip(context.Background())))
		})
	}
}

func TestDescribeHTTPError(t *testing.T) {
	wrap := func(err error) error { return &url.Error{Op: "Get", URL: "https://icanhazip.com", Err: err} }

	tests := []struct {
		err      error
		expected string
	}{
		{wrap(context.DeadlineExceeded), "timed out"},
		{wrap(&net.DNSError{Err: "no such host", Name: "icanhazip.com"}), "can't resolve icanhazip.com"},
		{wrap(&tls.CertificateVerificationError{Err: errors.New("x509: unknown authority")}), "TLS certificate verification failed - missing CA certificates?"},
		{wrap(&net.OpError{Op: "dial", Err: errors.New("network is unreachable")}), "no internet connection"},
		{wrap(fmt.Errorf("something else")), `Get "https://icanhazip.com": something else`},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.expected, describeHTTPError(tt.err).Error())
	}
}
