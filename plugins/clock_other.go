//go:build !linux

package plugins

// isClockSynchronized is implemented only on Linux.
func isClockSynchronized() (synchronized bool, ok bool) {
	return false, false
}
