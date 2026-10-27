package plugins

import (
	"cmp"
	"context"
	"fmt"
	"runtime"
	"slices"
	"time"

	"github.com/Tonyfilla/go-humanize"
	"github.com/gdubicki/dynamotd/dynamotd"
	"github.com/shirou/gopsutil/v4/load"
	"github.com/shirou/gopsutil/v4/mem"
	"github.com/shirou/gopsutil/v4/process"
)

type topProcessesSettings struct {
	// show the top processes even if the load and memory usage are below the thresholds
	Always bool `yaml:"always"`
	// show the top processes by CPU usage when the 1 minute load average per CPU core reaches this
	Load float64 `yaml:"load"`
	// show the top processes by memory usage when the RAM usage reaches this, in %
	Memory float64 `yaml:"memory"`
	// how many processes to show
	Count int `yaml:"count"`
}

func (settings topProcessesSettings) Validate() error {
	if settings.Count < 1 {
		return fmt.Errorf("count must be at least 1")
	}
	return nil
}

func init() {
	dynamotd.Register(topProcessesLabel, topProcessesSettings{Load: 0.9, Memory: 90, Count: 3}, TopProcesses, "topprocesses")
}

const topProcessesLabel = "Top processes"

// how long to measure the CPU usage of the processes for
const cpuSamplingTime = 250 * time.Millisecond

type processSample struct {
	pid        int32
	name       string
	cpuPercent float64
	rss        uint64
}

// for mocking in tests
var getLoadPerCore = func(ctx context.Context) (float64, error) {
	avg, err := load.AvgWithContext(ctx)
	if err != nil {
		return 0, err
	}
	return avg.Load1 / float64(runtime.NumCPU()), nil
}

var getMemoryUsedPercent = func(ctx context.Context) (float64, error) {
	memory, err := mem.VirtualMemoryWithContext(ctx)
	if err != nil {
		return 0, err
	}
	// the same as in the memory row
	return float64(memory.Total-memory.Available) / float64(memory.Total) * 100, nil
}

var sampleProcesses = func(ctx context.Context, measureCPU bool) ([]processSample, error) {
	processes, err := process.ProcessesWithContext(ctx)
	if err != nil {
		return nil, err
	}

	cpuTime := func(p *process.Process) (float64, bool) {
		times, err := p.TimesWithContext(ctx)
		if err != nil {
			return 0, false
		}
		return times.User + times.System, true
	}

	cpuTimesBefore := map[int32]float64{}
	start := time.Now()
	if measureCPU {
		for _, p := range processes {
			if t, ok := cpuTime(p); ok {
				cpuTimesBefore[p.Pid] = t
			}
		}
		select {
		case <-time.After(cpuSamplingTime):
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
	elapsed := time.Since(start).Seconds()

	var samples []processSample
	for _, p := range processes {
		// the processes may have ended in the meantime, or we may not have access to them
		memoryInfo, err := p.MemoryInfoWithContext(ctx)
		if err != nil {
			continue
		}
		name, err := p.NameWithContext(ctx)
		if err != nil {
			continue
		}
		sample := processSample{pid: p.Pid, name: name, rss: memoryInfo.RSS}
		if before, ok := cpuTimesBefore[p.Pid]; ok {
			if after, ok := cpuTime(p); ok {
				// like in top, it's the % of a single CPU core, so it can be more than 100%
				sample.cpuPercent = (after - before) / elapsed * 100
			}
		}
		samples = append(samples, sample)
	}
	return samples, nil
}

// TopProcesses shows the processes using the most CPU when the load is high,
// and the ones using the most memory when the RAM usage is high.
func TopProcesses(ctx context.Context, settings topProcessesSettings) []dynamotd.Row {
	showCPU, showMemory := settings.Always, settings.Always
	if loadPerCore, err := getLoadPerCore(ctx); err == nil && loadPerCore >= settings.Load {
		showCPU = true
	}
	if memoryUsedPercent, err := getMemoryUsedPercent(ctx); err == nil && memoryUsedPercent >= settings.Memory {
		showMemory = true
	}
	if !showCPU && !showMemory {
		return nil
	}

	samples, err := sampleProcesses(ctx, showCPU)
	if err != nil {
		return []dynamotd.Row{dynamotd.ErrorRow(topProcessesLabel, fmt.Errorf("error getting processes: %w", err))}
	}

	var rows []dynamotd.Row
	if showCPU {
		rows = append(rows, topProcessesRow("Top CPU", samples, settings.Count,
			func(s processSample) float64 { return s.cpuPercent },
			func(s processSample) string { return fmt.Sprintf("%.0f%%", s.cpuPercent) }))
	}
	if showMemory {
		rows = append(rows, topProcessesRow("Top RAM", samples, settings.Count,
			func(s processSample) float64 { return float64(s.rss) },
			func(s processSample) string { return humanize.IBytesCustomCeil(s.rss, 2) }))
	}
	return rows
}

func topProcessesRow(label string, samples []processSample, count int,
	usage func(processSample) float64, formatUsage func(processSample) string) dynamotd.Row {

	sorted := slices.Clone(samples)
	slices.SortStableFunc(sorted, func(a, b processSample) int { return cmp.Compare(usage(b), usage(a)) })

	var value []dynamotd.ColorString
	for i, sample := range sorted[:min(count, len(sorted))] {
		if i > 0 {
			value = append(value, dynamotd.ValueDescription(", "))
		}
		value = append(value,
			dynamotd.ValueDescription(fmt.Sprintf("%s (%d) ", sample.name, sample.pid)),
			dynamotd.ValueNeutral(formatUsage(sample)),
		)
	}

	return dynamotd.Row{
		Label: dynamotd.SingleColorLabel(label),
		Value: dynamotd.ToColorText(value...),
	}
}
