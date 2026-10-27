package plugins

import (
	"context"
	"fmt"
	"runtime"
	"time"

	"github.com/gdubicki/dynamotd/dynamotd"
	"github.com/shirou/gopsutil/v4/cpu"
)

type cpuUsageSettings struct {
	// thresholds for the time spent waiting for I/O, in %
	IOWait dynamotd.Thresholds `yaml:"iowait"`
	// thresholds for the time stolen by the hypervisor from this VM, in %
	Steal dynamotd.Thresholds `yaml:"steal"`
}

func (settings cpuUsageSettings) Validate() error {
	if err := settings.IOWait.Validate(); err != nil {
		return fmt.Errorf("iowait: %w", err)
	}
	if err := settings.Steal.Validate(); err != nil {
		return fmt.Errorf("steal: %w", err)
	}
	return nil
}

func init() {
	dynamotd.Register(cpuUsageLabel, cpuUsageSettings{
		IOWait: dynamotd.Thresholds{Warning: 10, Critical: 25},
		Steal:  dynamotd.Thresholds{Warning: 5, Critical: 15},
	}, CPUUsage, "cpu")
}

const cpuUsageLabel = "CPU usage"

// iowait and steal are reported only on Linux
var showIOWaitAndSteal = runtime.GOOS == "linux"

// for mocking in tests
var sampleCPUTimes = func(ctx context.Context) (before, after cpu.TimesStat, err error) {
	sample := func() (cpu.TimesStat, error) {
		times, err := cpu.TimesWithContext(ctx, false)
		if err != nil {
			return cpu.TimesStat{}, err
		}
		if len(times) == 0 {
			return cpu.TimesStat{}, fmt.Errorf("no CPU times")
		}
		return times[0], nil
	}

	if before, err = sample(); err != nil {
		return
	}
	select {
	case <-time.After(cpuSamplingTime):
	case <-ctx.Done():
		err = ctx.Err()
		return
	}
	after, err = sample()
	return
}

// CPUUsage shows how the CPU time is spent - in user space, kernel, waiting for I/O
// or stolen by the hypervisor, as the load average alone doesn't tell that.
func CPUUsage(ctx context.Context, settings cpuUsageSettings) []dynamotd.Row {
	before, after, err := sampleCPUTimes(ctx)
	if err != nil {
		return []dynamotd.Row{dynamotd.ErrorRow(cpuUsageLabel, fmt.Errorf("error getting CPU times: %w", err))}
	}

	user := (after.User + after.Nice) - (before.User + before.Nice)
	system := (after.System + after.Irq + after.Softirq) - (before.System + before.Irq + before.Softirq)
	ioWait := after.Iowait - before.Iowait
	steal := after.Steal - before.Steal
	idle := after.Idle - before.Idle
	total := user + system + ioWait + steal + idle
	if total <= 0 {
		return []dynamotd.Row{dynamotd.ErrorRow(cpuUsageLabel, fmt.Errorf("no CPU time passed"))}
	}
	percent := func(value float64) float64 { return value / total * 100 }

	value := []dynamotd.ColorString{
		dynamotd.ValueNeutral(fmt.Sprintf("%.0f%%", percent(user))),
		dynamotd.ValueDescription(" user, "),
		dynamotd.ValueNeutral(fmt.Sprintf("%.0f%%", percent(system))),
		dynamotd.ValueDescription(" system"),
	}
	if showIOWaitAndSteal {
		value = append(value,
			dynamotd.ValueDescription(", "),
			dynamotd.ColorString{Color: settings.IOWait.Color(percent(ioWait)), Text: fmt.Sprintf("%.0f%%", percent(ioWait))},
			dynamotd.ValueDescription(" iowait, "),
			dynamotd.ColorString{Color: settings.Steal.Color(percent(steal)), Text: fmt.Sprintf("%.0f%%", percent(steal))},
			dynamotd.ValueDescription(" steal"),
		)
	}

	return []dynamotd.Row{{
		Label: dynamotd.SingleColorLabel(cpuUsageLabel),
		Value: dynamotd.ToColorText(value...),
	}}
}
