package plugins

import (
	"context"
	"testing"

	"github.com/gdubicki/dynamotd/dynamotd"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func mockCPUTimes(t *testing.T, linux bool, before, after cpu.TimesStat) {
	origSample, origShow := sampleCPUTimes, showIOWaitAndSteal
	t.Cleanup(func() { sampleCPUTimes, showIOWaitAndSteal = origSample, origShow })

	showIOWaitAndSteal = linux
	sampleCPUTimes = func(context.Context) (cpu.TimesStat, cpu.TimesStat, error) { return before, after, nil }
}

var defaultCPUUsageSettings = cpuUsageSettings{
	IOWait: cpuTimeThresholds{Show: 5, Thresholds: dynamotd.Thresholds{Warning: 10, Critical: 25}},
	Steal:  cpuTimeThresholds{Show: 2, Thresholds: dynamotd.Thresholds{Warning: 5, Critical: 15}},
}

func TestCPUUsageLinux(t *testing.T) {
	before := cpu.TimesStat{User: 100, Nice: 10, System: 50, Irq: 1, Softirq: 1, Iowait: 20, Steal: 5, Idle: 1000}
	// +15 user, +5 system, +30 iowait, +6 steal, +44 idle = 100
	after := cpu.TimesStat{User: 110, Nice: 15, System: 53, Irq: 2, Softirq: 2, Iowait: 50, Steal: 11, Idle: 1044}
	mockCPUTimes(t, true, before, after)

	rows := CPUUsage(context.Background(), defaultCPUUsageSettings)

	require.Len(t, rows, 1)
	assert.Equal(t, "20% busy (15% user, 5% kernel), 30% waiting for disk, 6% stolen by hypervisor", rows[0].Value.String())
	parts := rows[0].Value.Parts()
	assert.Equal(t, dynamotd.ValueCriticalColor, parts[7].Color, "waiting for disk")
	assert.Equal(t, dynamotd.ValueWarningColor, parts[10].Color, "stolen by hypervisor")
}

func TestCPUUsageLinuxLowIOWaitAndSteal(t *testing.T) {
	// 4% iowait and 1% steal are below the show thresholds
	mockCPUTimes(t, true, cpu.TimesStat{}, cpu.TimesStat{User: 30, System: 10, Iowait: 4, Steal: 1, Idle: 55})

	rows := CPUUsage(context.Background(), defaultCPUUsageSettings)

	require.Len(t, rows, 1)
	assert.Equal(t, "40% busy (30% user, 10% kernel)", rows[0].Value.String())
}

func TestCPUUsageLinuxOnlyStealShown(t *testing.T) {
	mockCPUTimes(t, true, cpu.TimesStat{}, cpu.TimesStat{User: 30, System: 10, Iowait: 1, Steal: 3, Idle: 56})

	rows := CPUUsage(context.Background(), defaultCPUUsageSettings)

	require.Len(t, rows, 1)
	assert.Equal(t, "40% busy (30% user, 10% kernel), 3% stolen by hypervisor", rows[0].Value.String())
	assert.Equal(t, dynamotd.ValueOkColor, rows[0].Value.Parts()[7].Color)
}

func TestCPUUsageMacOS(t *testing.T) {
	mockCPUTimes(t, false, cpu.TimesStat{}, cpu.TimesStat{User: 30, System: 10, Idle: 60})

	rows := CPUUsage(context.Background(), defaultCPUUsageSettings)

	require.Len(t, rows, 1)
	assert.Equal(t, "40% busy (30% user, 10% kernel)", rows[0].Value.String())
}
