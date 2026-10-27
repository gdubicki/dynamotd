//go:build !linux

package plugins

import "errors"

// readKmsg is implemented only on Linux.
func readKmsg() ([]string, error) {
	return nil, errors.New("not supported")
}
