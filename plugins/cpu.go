package plugins

import (
	"context"
	"errors"
	"fmt"
	"runtime"
	"time"

	"github.com/gdubicki/dynamotd/dynamotd"
	"github.com/shirou/gopsutil/v4/cpu"
)

type cpuUsageSettings struct {
	// thresholds for the time spent waiting for I/O, in %
	IOWait cpuTimeThresholds `yaml:"iowait"`
	// thresholds for the time stolen by the hypervisor from this VM, in %
	Steal cpuTimeThresholds `yaml:"steal"`
}

type cpuTimeThresholds struct {
	// the value is shown only if it's at least this, as low values are normal and just make the line longer
	Show                float64 `yaml:"show"`
	dynamotd.Thresholds `yaml:",inline"`
}

func (thresholds cpuTimeThresholds) Validate() error {
	if thresholds.Show > thresholds.Warning {
		return fmt.Errorf("show threshold (%v) is higher than the warning one (%v)", thresholds.Show, thresholds.Warning)
	}
	return thresholds.Thresholds.Validate()
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
		IOWait: cpuTimeThresholds{Show: 5, Thresholds: dynamotd.Thresholds{Warning: 10, Critical: 25}},
		Steal:  cpuTimeThresholds{Show: 2, Thresholds: dynamotd.Thresholds{Warning: 5, Critical: 15}},
	}, CPUUsage, "cpu")
}

const cpuUsageLabel = "CPU usage"

// iowait and steal are reported only on Linux
var showIOWaitAndSteal = runtime.GOOS == "linux"

// for mocking in tests
var sampleCPUTimes = func(ctx context.Context) (before, after cpu.TimesStat, err error) {
	// sum the per-CPU times, as on macOS the total ones are sometimes not updated for a second or more
	sample := func() (cpu.TimesStat, error) {
		perCPU, err := cpu.TimesWithContext(ctx, true)
		if err != nil {
			return cpu.TimesStat{}, err
		}
		if len(perCPU) == 0 {
			return cpu.TimesStat{}, fmt.Errorf("no CPU times")
		}
		var total cpu.TimesStat
		for _, t := range perCPU {
			total.User += t.User
			total.Nice += t.Nice
			total.System += t.System
			total.Irq += t.Irq
			total.Softirq += t.Softirq
			total.Iowait += t.Iowait
			total.Steal += t.Steal
			total.Idle += t.Idle
		}
		return total, nil
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
	if errors.Is(err, context.DeadlineExceeded) {
		return []dynamotd.Row{dynamotd.TimedOutRow(cpuUsageLabel)}
	}
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

	// f.e. "25% busy (16% user, 9% kernel), 30% waiting for disk, 3% stolen by hypervisor"
	value := []dynamotd.ColorString{
		dynamotd.ValueNeutral(fmt.Sprintf("%.0f%%", percent(user+system))),
		dynamotd.ValueDescription(" busy ("),
		dynamotd.ValueNeutral(fmt.Sprintf("%.0f%%", percent(user))),
		dynamotd.ValueDescription(" user, "),
		dynamotd.ValueNeutral(fmt.Sprintf("%.0f%%", percent(system))),
		dynamotd.ValueDescription(" kernel)"),
	}
	if showIOWaitAndSteal {
		for _, cpuTime := range []struct {
			percent     float64
			thresholds  cpuTimeThresholds
			description string
		}{
			{percent(ioWait), settings.IOWait, " waiting for disk"},
			{percent(steal), settings.Steal, " stolen by hypervisor"},
		} {
			if cpuTime.percent < cpuTime.thresholds.Show {
				continue
			}
			value = append(value,
				dynamotd.ValueDescription(", "),
				dynamotd.ColorString{Color: cpuTime.thresholds.Color(cpuTime.percent), Text: fmt.Sprintf("%.0f%%", cpuTime.percent)},
				dynamotd.ValueDescription(cpuTime.description),
			)
		}
	}

	return []dynamotd.Row{{
		Label: dynamotd.SingleColorLabel(cpuUsageLabel),
		Value: dynamotd.ToColorText(value...),
	}}
}
