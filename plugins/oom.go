package plugins

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gdubicki/dynamotd/dynamotd"
	"github.com/shirou/gopsutil/v4/host"
)

// OOM kills more recent than this are shown as critical, older ones as warnings,
// as they have likely already been dealt with
const recentOOMKill = 24 * time.Hour

type oomKill struct {
	time    time.Time
	process string
}

// for mocking in tests
var readKernelLog = readKmsg
var getBootTime = host.BootTime
var now = time.Now

// matches both "Out of memory: Killed process 1234 (java) ..." and
// "Memory cgroup out of memory: Killed process 1234 (java) ..."
var killedProcessRegexp = regexp.MustCompile(`Killed process \d+ \(([^)]+)\)`)

// lastOOMKill returns the last OOM kill found in the kernel log. It may be not found if the log
// is not readable (it requires root) or if the kernel's log buffer has been overwritten since.
func lastOOMKill() (oomKill, bool) {
	records, err := readKernelLog()
	if err != nil {
		return oomKill{}, false
	}
	bootTime, err := getBootTime()
	if err != nil {
		return oomKill{}, false
	}
	return findLastOOMKill(records, time.Unix(int64(bootTime), 0))
}

// findLastOOMKill finds the last OOM kill in the /dev/kmsg records, which are like:
// "6,1234,5678901234,-;Out of memory: Killed process 4321 (java) total-vm:..."
// where the 3rd field is the time since boot in µs.
func findLastOOMKill(records []string, bootTime time.Time) (oomKill, bool) {
	for i := len(records) - 1; i >= 0; i-- {
		header, message, found := strings.Cut(records[i], ";")
		if !found {
			continue
		}
		match := killedProcessRegexp.FindStringSubmatch(message)
		if match == nil {
			continue
		}
		fields := strings.Split(header, ",")
		if len(fields) < 3 {
			continue
		}
		sinceBoot, err := strconv.ParseInt(fields[2], 10, 64)
		if err != nil {
			continue
		}
		return oomKill{time: bootTime.Add(time.Duration(sinceBoot) * time.Microsecond), process: match[1]}, true
	}
	return oomKill{}, false
}

// oomKillsValue describes the OOM kills since boot, f.e.
// "3 OOM kills since boot, last 2 days ago (java, 2026-10-25 14:03)"
func oomKillsValue(count uint64) []dynamotd.ColorString {
	unit := "OOM kill"
	if count > 1 {
		unit += "s"
	}

	last, found := lastOOMKill()
	countColor := dynamotd.ValueCriticalColor
	if found && now().Sub(last.time) > recentOOMKill {
		countColor = dynamotd.ValueWarningColor
	}

	value := []dynamotd.ColorString{
		{Color: countColor, Text: fmt.Sprintf("%d %s", count, unit)},
		dynamotd.ValueDescription(" since boot"),
	}
	if found {
		value = append(value,
			dynamotd.ValueDescription(", last "),
			dynamotd.ColorString{Color: countColor, Text: ago(now().Sub(last.time))},
			dynamotd.ValueDescription(fmt.Sprintf(" (%s, %s)", last.process, last.time.Local().Format("2006-01-02 15:04"))),
		)
	}
	return value
}

// ago returns a rough, human-readable description of how long ago something happened.
func ago(duration time.Duration) string {
	plural := func(n int, unit string) string {
		if n == 1 {
			return fmt.Sprintf("1 %s ago", unit)
		}
		return fmt.Sprintf("%d %ss ago", n, unit)
	}
	switch {
	case duration < time.Minute:
		return "just now"
	case duration < time.Hour:
		return plural(int(duration.Minutes()), "minute")
	case duration < 48*time.Hour:
		return plural(int(duration.Hours()), "hour")
	default:
		return plural(int(duration.Hours()/24), "day")
	}
}
