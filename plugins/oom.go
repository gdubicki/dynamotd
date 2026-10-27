package plugins

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gdubicki/dynamotd/dynamotd"
	"github.com/shirou/gopsutil/v4/host"
)

func init() {
	dynamotd.Register(oomKillsLabel, dynamotd.NoSettings{}, OOMKills, "oomkills")
}

const oomKillsLabel = "OOM kills"

// OOM kills more recent than this are shown as critical, older ones as warnings,
// as they have likely already been dealt with
const recentOOMKill = 24 * time.Hour

type oomKill struct {
	time    time.Time
	process string
}

// lastOOMKillInfo is what is known about the last OOM kill from the kernel log
type lastOOMKillInfo struct {
	// the last kill, if it's still in the kernel log
	kill *oomKill
	// otherwise, the time of the oldest record in the kernel log - the last kill happened before it
	before time.Time
}

// for mocking in tests
var vmstatPath = "/proc/vmstat"
var readKernelLog = readKmsg
var getBootTime = host.BootTime
var now = time.Now

// matches both "Out of memory: Killed process 1234 (java) ..." and
// "Memory cgroup out of memory: Killed process 1234 (java) ..."
var killedProcessRegexp = regexp.MustCompile(`Killed process \d+ \(([^)]+)\)`)

// lastOOMKill returns what is known about the last OOM kill from the kernel log, which is not readable
// without root. The kill may be no longer in the kernel's log buffer if it has been overwritten since,
// but then we know at least that it happened before the oldest record in it.
func lastOOMKill() (lastOOMKillInfo, bool) {
	records, err := readKernelLog()
	if err != nil {
		return lastOOMKillInfo{}, false
	}
	bootTime, err := getBootTime()
	if err != nil {
		return lastOOMKillInfo{}, false
	}
	return findLastOOMKill(records, time.Unix(int64(bootTime), 0))
}

// findLastOOMKill finds the last OOM kill in the /dev/kmsg records, which are like:
// "6,1234,5678901234,-;Out of memory: Killed process 4321 (java) total-vm:..."
// where the 3rd field is the time since boot in µs.
func findLastOOMKill(records []string, bootTime time.Time) (lastOOMKillInfo, bool) {
	for i := len(records) - 1; i >= 0; i-- {
		header, message, found := strings.Cut(records[i], ";")
		if !found {
			continue
		}
		match := killedProcessRegexp.FindStringSubmatch(message)
		if match == nil {
			continue
		}
		if recordTime, ok := kmsgRecordTime(header, bootTime); ok {
			return lastOOMKillInfo{kill: &oomKill{time: recordTime, process: match[1]}}, true
		}
	}

	// not found, so it happened before the oldest record
	for _, record := range records {
		header, _, _ := strings.Cut(record, ";")
		if recordTime, ok := kmsgRecordTime(header, bootTime); ok {
			return lastOOMKillInfo{before: recordTime}, true
		}
	}
	return lastOOMKillInfo{}, false
}

func kmsgRecordTime(header string, bootTime time.Time) (time.Time, bool) {
	fields := strings.Split(header, ",")
	if len(fields) < 3 {
		return time.Time{}, false
	}
	sinceBoot, err := strconv.ParseInt(fields[2], 10, 64)
	if err != nil {
		return time.Time{}, false
	}
	return bootTime.Add(time.Duration(sinceBoot) * time.Microsecond), true
}

// OOMKills shows the number of processes killed by the OOM killer since boot, if there were any,
// and when the last one was killed, f.e. "3 since boot, last 2 days ago (java, 2026-10-25 14:03)",
// or if it's no longer in the kernel log: "3 since boot, last over 1 day ago (before 2026-10-26 10:56)".
func OOMKills(_ context.Context, _ dynamotd.NoSettings) []dynamotd.Row {
	// it's not available on macOS and old Linux kernels
	count, err := getOOMKills()
	if err != nil || count == 0 {
		return nil
	}

	last, found := lastOOMKill()
	// the last kill happened at or before this time
	var lastTime time.Time
	if found && last.kill != nil {
		lastTime = last.kill.time
	} else if found {
		lastTime = last.before
	}

	color := dynamotd.ValueCriticalColor
	if found && now().Sub(lastTime) > recentOOMKill {
		color = dynamotd.ValueWarningColor
	}

	value := []dynamotd.ColorString{
		{Color: color, Text: fmt.Sprintf("%d", count)},
		dynamotd.ValueDescription(" since boot"),
	}
	switch {
	case found && last.kill != nil:
		value = append(value,
			dynamotd.ValueDescription(", last "),
			dynamotd.ColorString{Color: color, Text: ago(now().Sub(lastTime))},
			dynamotd.ValueDescription(fmt.Sprintf(" (%s, %s)", last.kill.process, formatOOMTime(lastTime))),
		)
	case found:
		value = append(value,
			dynamotd.ValueDescription(", last "),
			dynamotd.ColorString{Color: color, Text: "over " + ago(now().Sub(lastTime))},
			dynamotd.ValueDescription(fmt.Sprintf(" (before %s)", formatOOMTime(lastTime))),
		)
	}

	return []dynamotd.Row{{
		Label: dynamotd.SingleColorLabel(oomKillsLabel),
		Value: dynamotd.ToColorText(value...),
	}}
}

func formatOOMTime(t time.Time) string {
	return t.Local().Format("2006-01-02 15:04")
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

// getOOMKills returns the number of processes killed by the OOM killer since boot.
func getOOMKills() (uint64, error) {
	file, err := os.Open(vmstatPath)
	if err != nil {
		return 0, err
	}
	defer func() { _ = file.Close() }()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		if value, found := strings.CutPrefix(scanner.Text(), "oom_kill "); found {
			return strconv.ParseUint(value, 10, 64)
		}
	}
	if err := scanner.Err(); err != nil {
		return 0, err
	}
	return 0, fmt.Errorf("no oom_kill in %s", vmstatPath)
}
