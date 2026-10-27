package plugins

import (
	"errors"

	"golang.org/x/sys/unix"
)

// readKmsg returns all the records currently in the kernel's log buffer. Reading it requires root
// (or CAP_SYSLOG) on most distros. It uses raw syscalls, as reading /dev/kmsg with the Go's os.File
// would block waiting for new records instead of returning when all the existing ones are read.
func readKmsg() ([]string, error) {
	fd, err := unix.Open("/dev/kmsg", unix.O_RDONLY|unix.O_NONBLOCK|unix.O_CLOEXEC, 0)
	if err != nil {
		return nil, err
	}
	defer func() { _ = unix.Close(fd) }()

	var records []string
	// each read returns a single record, which is at most this long
	buffer := make([]byte, 8192)
	for {
		n, err := unix.Read(fd, buffer)
		switch {
		case errors.Is(err, unix.EAGAIN):
			// no more records
			return records, nil
		case errors.Is(err, unix.EPIPE):
			// the record was overwritten while reading, skip it
			continue
		case err != nil:
			return nil, err
		}
		records = append(records, string(buffer[:n]))
	}
}
