package plugins

import (
	"errors"
	"testing"
	"time"

	"github.com/gdubicki/dynamotd/dynamotd"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var testBootTime = time.Date(2026, 10, 20, 12, 0, 0, 0, time.Local)

// mockKernelLog mocks the kernel log records, or makes it unreadable if records is nil
func mockKernelLog(t *testing.T, records []string) {
	origRead, origBoot, origNow := readKernelLog, getBootTime, now
	t.Cleanup(func() { readKernelLog, getBootTime, now = origRead, origBoot, origNow })

	readKernelLog = func() ([]string, error) {
		if records == nil {
			return nil, errors.New("permission denied")
		}
		return records, nil
	}
	getBootTime = func() (uint64, error) { return uint64(testBootTime.Unix()), nil }
}

func setNow(t *testing.T, timestamp time.Time) {
	now = func() time.Time { return timestamp }
}

var kernelLogWithOOMKills = []string{
	"6,100,1000000,-;Linux version 6.8.0-50-generic",
	// 1h after boot
	"3,200,3600000000,-;Out of memory: Killed process 4321 (java) total-vm:8000kB, anon-rss:4000kB",
	// 2 days after boot
	"3,300,172800000000,-;Memory cgroup out of memory: Killed process 999 (postgres) total-vm:100kB",
	"6,301,172800000001,-;oom_reaper: reaped process 999 (postgres), now anon-rss:0kB",
	"6,400,172900000000,-;some other message",
}

func TestFindLastOOMKill(t *testing.T) {
	last, found := findLastOOMKill(kernelLogWithOOMKills, testBootTime)

	require.True(t, found)
	require.NotNil(t, last.kill)
	assert.Equal(t, "postgres", last.kill.process)
	assert.Equal(t, testBootTime.Add(48*time.Hour), last.kill.time)

	// the kills are no longer in the kernel log, so they happened before its oldest record
	last, found = findLastOOMKill(kernelLogWithOOMKills[4:], testBootTime)
	require.True(t, found)
	assert.Nil(t, last.kill)
	assert.Equal(t, testBootTime.Add(172900*time.Second), last.before)

	_, found = findLastOOMKill(nil, testBootTime)
	assert.False(t, found)
}

func TestOOMKillsValue(t *testing.T) {
	lastKill := testBootTime.Add(48 * time.Hour)
	oldestRecord := testBootTime.Add(172900 * time.Second)

	tests := []struct {
		name     string
		records  []string
		now      time.Time
		expected string
		color    string
	}{
		{"recent", kernelLogWithOOMKills, lastKill.Add(3 * time.Hour),
			"2 OOM kills since boot, last 3 hours ago (postgres, " + lastKill.Format("2006-01-02 15:04") + ")", "critical"},
		{"old", kernelLogWithOOMKills, lastKill.Add(50 * time.Hour),
			"2 OOM kills since boot, last 2 days ago (postgres, " + lastKill.Format("2006-01-02 15:04") + ")", "warning"},
		{"kernel log not readable", nil, lastKill,
			"2 OOM kills since boot", "critical"},
		{"no longer in the kernel log, which starts over 24h ago", kernelLogWithOOMKills[4:], lastKill.Add(30 * time.Hour),
			"2 OOM kills since boot, last over 29 hours ago (before " + oldestRecord.Format("2006-01-02 15:04") + ")", "warning"},
		{"no longer in the kernel log, which starts within 24h", kernelLogWithOOMKills[4:], lastKill.Add(3 * time.Hour),
			"2 OOM kills since boot, last over 2 hours ago (before " + oldestRecord.Format("2006-01-02 15:04") + ")", "critical"},
		{"empty kernel log", []string{}, lastKill,
			"2 OOM kills since boot", "critical"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockKernelLog(t, tt.records)
			setNow(t, tt.now)

			value := dynamotd.ToColorText(oomKillsValue(2)...)

			assert.Equal(t, tt.expected, value.String())
			expectedColor := dynamotd.ValueCriticalColor
			if tt.color == "warning" {
				expectedColor = dynamotd.ValueWarningColor
			}
			assert.Equal(t, expectedColor, value.Parts()[0].Color)
		})
	}
}

func TestAgo(t *testing.T) {
	tests := map[time.Duration]string{
		30 * time.Second:                "just now",
		time.Minute:                     "1 minute ago",
		59 * time.Minute:                "59 minutes ago",
		time.Hour:                       "1 hour ago",
		47*time.Hour + 59*time.Minute:   "47 hours ago",
		48 * time.Hour:                  "2 days ago",
		30*24*time.Hour + 5*time.Minute: "30 days ago",
	}
	for duration, expected := range tests {
		assert.Equal(t, expected, ago(duration), duration.String())
	}
}
