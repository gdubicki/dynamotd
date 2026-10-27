package plugins

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

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
