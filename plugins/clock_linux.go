package plugins

import "golang.org/x/sys/unix"

// the kernel caps the maximum error at this value (16s, in µs) when the clock is not synchronized
const maxErrorWhenUnsynchronized = 16_000_000

// isClockSynchronized returns the kernel's view of whether the clock is synchronized, f.e. by NTP,
// the same way as systemd does (and shows in "timedatectl"). Reading it doesn't require any privileges.
func isClockSynchronized() (synchronized bool, ok bool) {
	var timex unix.Timex
	if _, err := unix.Adjtimex(&timex); err != nil {
		return false, false
	}
	// like systemd, ignore the STA_UNSYNC flag, as it may be set only to prevent the kernel from updating the RTC
	return timex.Maxerror < maxErrorWhenUnsynchronized, true
}
